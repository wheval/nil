package core

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/NilFoundation/nil/nil/go-ibft/messages"
	"github.com/NilFoundation/nil/nil/go-ibft/messages/proto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("go-ibft")

// Logger represents the logger behaviour
type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

// Messages represents the message managing behaviour
type Messages interface {
	// Messages modifiers //
	AddMessage(message *proto.IbftMessage)
	PruneByHeight(height uint64)

	SignalEvent(messageType proto.MessageType, view *proto.View)

	// Messages fetchers //
	GetValidMessages(
		view *proto.View,
		messageType proto.MessageType,
		isValid func(*proto.IbftMessage) bool,
	) []*proto.IbftMessage
	GetExtendedRCC(
		height uint64,
		isValidMessage func(message *proto.IbftMessage) bool,
		isValidRCC func(round uint64, msgs []*proto.IbftMessage) bool,
	) []*proto.IbftMessage
	GetMostRoundChangeMessages(minRound, height uint64) []*proto.IbftMessage

	// Messages subscription handlers //
	Subscribe(details messages.SubscriptionDetails) *messages.Subscription
	Unsubscribe(id messages.SubscriptionID)
}

const (
	round0Timeout   = 10 * time.Second
	roundFactorBase = float64(2)
)

var errTimeoutExpired = errors.New("round timeout expired")

// IBFT represents a single instance of the IBFT state machine
type IBFT struct {
	// log is the logger instance
	log Logger

	// state is the current IBFT node state
	state *state

	// messages is the message storage layer
	messages Messages

	// backend is the reference to the
	// Backend implementation
	backend Backend

	// transport is the reference to the
	// Transport implementation
	transport Transport

	// roundDone is the channel used for signalizing
	// consensus finalization upon a certain sequence
	roundDone chan struct{}

	// roundExpired is the channel used for signalizing
	// round changing events
	roundExpired chan struct{}

	// newProposal is the channel used for signalizing
	// when new proposals for a view greater than the current
	// one arrive
	newProposal chan newProposalEvent

	// roundCertificate is the channel used for signalizing
	// when a valid RCC for a greater round than the current
	// one is present
	roundCertificate chan uint64

	//	User configured additional timeout for each round of consensus
	additionalTimeout time.Duration

	// baseRoundTimeout is the base round timeout for each round of consensus
	baseRoundTimeout time.Duration

	// wg is a simple barrier used for synchronizing
	// state modification routines
	wg sync.WaitGroup

	// validatorManager keeps quorumSize and voting power information
	validatorManager *ValidatorManager

	// metricAttrs is the attributes to be used for metrics
	metricAttrs []attribute.KeyValue

	// roundHistogram is the histogram for round duration
	roundHistogram metric.Float64Histogram

	// sequenceHistogram is the histogram for sequence duration
	sequenceHistogram metric.Float64Histogram

	// prepareHistogram is the histogram for "prepare" stage duration
	prepareHistogram metric.Float64Histogram

	// newRoundHistogram is the histogram for "new round" stage duration
	newRoundHistogram metric.Float64Histogram

	// commitHistogram is the histogram for "commit" stage duration
	commitHistogram metric.Float64Histogram
}

// NewIBFTWithMetrics creates a new instance of the IBFT consensus protocol with enabled metrics
func NewIBFTWithMetrics(
	log Logger,
	backend Backend,
	transport Transport,
	attrs ...attribute.KeyValue,
) (*IBFT, error) {
	var err error
	ibft := NewIBFT(log, backend, transport)
	ibft.roundHistogram, err = meter.Float64Histogram(
		"round.duration", metric.WithDescription("Round duration"))
	if err != nil {
		return nil, err
	}
	ibft.sequenceHistogram, err = meter.Float64Histogram(
		"sequence.duration", metric.WithDescription("Sequence duration"),
	)
	if err != nil {
		return nil, err
	}
	ibft.prepareHistogram, err = meter.Float64Histogram(
		"prepare.duration", metric.WithDescription("Prepare stage duration"),
	)
	if err != nil {
		return nil, err
	}
	ibft.newRoundHistogram, err = meter.Float64Histogram(
		"new_round.duration", metric.WithDescription("New round stage duration"),
	)
	if err != nil {
		return nil, err
	}
	ibft.commitHistogram, err = meter.Float64Histogram(
		"commit.duration", metric.WithDescription("Commit stage duration"),
	)
	if err != nil {
		return nil, err
	}
	ibft.metricAttrs = attrs
	return ibft, nil
}

// NewIBFT creates a new instance of the IBFT consensus protocol
func NewIBFT(
	log Logger,
	backend Backend,
	transport Transport,
) *IBFT {
	return &IBFT{
		log:              log,
		backend:          backend,
		transport:        transport,
		messages:         messages.NewMessages(),
		roundDone:        make(chan struct{}),
		roundExpired:     make(chan struct{}),
		newProposal:      make(chan newProposalEvent),
		roundCertificate: make(chan uint64),
		state: &state{
			view: &proto.View{
				Height: 0,
				Round:  0,
			},
			seals:        make([]*messages.CommittedSeal, 0),
			roundStarted: false,
			name:         newRound,
		},
		baseRoundTimeout: round0Timeout,
		validatorManager: NewValidatorManager(backend, log),
	}
}

// setMeasurementTime function set duration to gauge
func (i *IBFT) setMeasurementTime(ctx context.Context, hist metric.Float64Histogram, startTime time.Time) {
	if hist == nil {
		return
	}
	hist.Record(ctx, float64(time.Since(startTime).Seconds()), metric.WithAttributes(i.metricAttrs...))
}

// startRoundTimer starts the exponential round timer, based on the
// passed in round number
func (i *IBFT) startRoundTimer(ctx context.Context, round uint64) {
	defer i.wg.Done()

	startTime := time.Now()

	roundTimeout := getRoundTimeout(i.baseRoundTimeout, i.additionalTimeout, round)

	//	Create a new timer instance
	timer := time.NewTimer(roundTimeout)

	select {
	case <-ctx.Done():
		i.setMeasurementTime(ctx, i.roundHistogram, startTime)
		// Stop signal received, stop the timer
		timer.Stop()
	case <-timer.C:
		// Timer expired, alert the round change channel to move
		// to the next round
		i.signalRoundExpired(ctx)
	}
}

// signalRoundExpired notifies the sequence routine (RunSequence) that it
// should move to a new round. The quit channel is used to abort this call
// if another routine has already signaled a round change request.
func (i *IBFT) signalRoundExpired(ctx context.Context) {
	select {
	case i.roundExpired <- struct{}{}:
	case <-ctx.Done():
	}
}

// signalRoundDone notifies the sequence routine (RunSequence) that the
// consensus sequence is finished
func (i *IBFT) signalRoundDone(ctx context.Context) {
	select {
	case i.roundDone <- struct{}{}:
	case <-ctx.Done():
	}
}

// signalNewRCC notifies the sequence routine (RunSequence) that
// a valid Round Change Certificate for a higher round appeared
func (i *IBFT) signalNewRCC(ctx context.Context, round uint64) {
	select {
	case i.roundCertificate <- round:
	case <-ctx.Done():
	}
}

type newProposalEvent struct {
	proposalMessage *proto.IbftMessage
	round           uint64
}

// signalNewProposal notifies the sequence routine (RunSequence) that
// a valid proposal for a higher round appeared
func (i *IBFT) signalNewProposal(ctx context.Context, event newProposalEvent) {
	select {
	case i.newProposal <- event:
	case <-ctx.Done():
	}
}

// watchForFutureProposal listens for new proposal messages
// that are intended for higher rounds
func (i *IBFT) watchForFutureProposal(ctx context.Context) {
	var (
		view      = i.state.getView()
		height    = view.GetHeight()
		nextRound = view.GetRound() + 1

		sub = i.subscribe(
			messages.SubscriptionDetails{
				MessageType: proto.MessageType_PREPREPARE,
				View: &proto.View{
					Height: height,
					Round:  nextRound,
				},
				HasMinRound: true,
			})
	)

	defer func() {
		i.messages.Unsubscribe(sub.ID)

		i.wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case round := <-sub.SubCh:
			proposal := i.handlePrePrepare(&proto.View{Height: height, Round: round})
			if proposal == nil {
				continue
			}

			// Extract the proposal
			i.signalNewProposal(
				ctx,
				newProposalEvent{proposal, round},
			)

			return
		}
	}
}

// watchForRoundChangeCertificates is a routine that waits
// for future valid Round Change Certificates that could
// trigger a round hop
func (i *IBFT) watchForRoundChangeCertificates(ctx context.Context) {
	defer i.wg.Done()

	var (
		view   = i.state.getView()
		height = view.GetHeight()
		round  = view.GetRound()

		sub = i.subscribe(messages.SubscriptionDetails{
			MessageType: proto.MessageType_ROUND_CHANGE,
			View: &proto.View{
				Height: height,
				Round:  round + 1, // only for higher rounds
			},
			HasMinRound: true,
		})
	)

	defer i.messages.Unsubscribe(sub.ID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-sub.SubCh:
			rcc := i.handleRoundChangeMessage(
				&proto.View{
					Height: height,
					Round:  round,
				},
			)
			if rcc == nil {
				continue
			}

			newRound := rcc.GetRoundChangeMessages()[0].GetView().GetRound()

			//	we received a valid RCC for a higher round
			i.signalNewRCC(ctx, newRound)

			return
		}
	}
}

// RunSequence runs the IBFT sequence for the specified height
func (i *IBFT) RunSequence(ctx context.Context, h uint64) {
	startTime := time.Now()

	// Set the starting state data
	i.state.reset(h)

	if err := i.validatorManager.Init(h); err != nil {
		i.log.Error("failed to run sequence - validator manager init", "height", h, "error", err)

		return
	}

	// Prune messages for older heights
	i.messages.PruneByHeight(h)

	i.log.Info("sequence started", "height", h)
	defer i.log.Info("sequence done", "height", h)
	defer i.setMeasurementTime(ctx, i.sequenceHistogram, startTime)

	for {
		view := i.state.getView()

		i.log.Info("round started", "round", view.GetRound())

		currentRound := view.GetRound()
		ctxRound, cancelRound := context.WithCancel(ctx)

		i.wg.Add(4)

		// Start the round timer worker
		go i.startRoundTimer(ctxRound, currentRound)

		//	Jump round on proposals from higher rounds
		go i.watchForFutureProposal(ctxRound)

		//	Jump round on certificates
		go i.watchForRoundChangeCertificates(ctxRound)

		// Start the state machine worker
		go i.startRound(ctxRound)

		teardown := func() {
			cancelRound()
			i.wg.Wait()
		}

		select {
		case ev := <-i.newProposal:
			teardown()
			i.log.Info("received future proposal", "round", ev.round)

			i.moveToNewRound(ev.round)
			i.acceptProposal(ev.proposalMessage)
			i.state.setRoundStarted(true)
			i.sendPrepareMessage(view)
		case round := <-i.roundCertificate:
			teardown()
			i.log.Info("received future RCC", "round", round)

			i.moveToNewRound(round)
		case <-i.roundExpired:
			teardown()
			i.log.Info("round timeout expired", "round", currentRound)

			newRound := currentRound + 1
			i.moveToNewRound(newRound)

			i.sendRoundChangeMessage(h, newRound)
		case <-i.roundDone:
			// The consensus cycle for the block height is finished.
			// Stop all running worker threads
			teardown()
			i.insertBlock()

			return
		case <-ctxRound.Done():
			teardown()
			i.log.Debug("sequence cancelled")

			return
		}
	}
}

// startRound runs the state machine loop for the current round
func (i *IBFT) startRound(ctx context.Context) {
	// Register this worker thread with the barrier
	defer i.wg.Done()

	i.state.newRound()

	var (
		id   = i.backend.ID()
		view = i.state.getView()
	)

	// Check if any block needs to be proposed
	if i.backend.IsProposer(id, view.GetHeight(), view.GetRound()) {
		i.log.Info("we are the proposer")

		proposalMessage := i.buildProposal(ctx, view)
		if proposalMessage == nil {
			i.log.Error("unable to build proposal")

			return
		}

		i.acceptProposal(proposalMessage)
		i.log.Debug("block proposal accepted")

		i.sendPreprepareMessage(proposalMessage)

		i.log.Debug("pre-prepare message multicasted")
	}

	i.runStates(ctx)
}

// waitForRCC waits for valid RCC for the specified height and round
func (i *IBFT) waitForRCC(
	ctx context.Context,
	height,
	round uint64,
) *proto.RoundChangeCertificate {
	var (
		view = &proto.View{
			Height: height,
			Round:  round,
		}

		sub = i.subscribe(
			messages.SubscriptionDetails{
				MessageType: proto.MessageType_ROUND_CHANGE,
				View:        view,
			},
		)
	)

	defer i.messages.Unsubscribe(sub.ID)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-sub.SubCh:
			rcc := i.handleRoundChangeMessage(view)
			if rcc == nil {
				continue
			}

			return rcc
		}
	}
}

// handleRoundChangeMessage validates the round change message
// and constructs a RCC if possible
func (i *IBFT) handleRoundChangeMessage(view *proto.View) *proto.RoundChangeCertificate {
	var (
		height              = view.GetHeight()
		hasAcceptedProposal = i.state.getProposal() != nil
	)

	isValidMsgFn := func(msg *proto.IbftMessage) bool {
		proposal := messages.ExtractLastPreparedProposal(msg)
		certificate := messages.ExtractLatestPC(msg)

		// Check if the prepared certificate is valid
		if !i.validPC(certificate, msg.GetView().GetRound(), height) {
			return false
		}

		// Make sure the certificate matches the proposal
		return i.proposalMatchesCertificate(proposal, certificate)
	}

	isValidRCCFn := func(round uint64, msgs []*proto.IbftMessage) bool {
		// In case of that ROUND-CHANGE message's round match validator's round
		// Accept such messages only if the validator has not accepted a proposal at the round
		if round == view.GetRound() && hasAcceptedProposal {
			return false
		}

		return i.hasQuorumByMsgType(msgs, proto.MessageType_ROUND_CHANGE)
	}

	extendedRCC := i.messages.GetExtendedRCC(
		height,
		isValidMsgFn,
		isValidRCCFn,
	)

	if extendedRCC == nil {
		return nil
	}

	return &proto.RoundChangeCertificate{
		RoundChangeMessages: extendedRCC,
	}
}

// proposalMatchesCertificate checks a prepared certificate
// against a proposal
func (i *IBFT) proposalMatchesCertificate(
	proposal *proto.Proposal,
	certificate *proto.PreparedCertificate,
) bool {
	// Both the certificate and proposal need to be set
	if proposal == nil && certificate == nil {
		return true
	}

	// If the proposal is set, the certificate also must be set
	if certificate == nil {
		return false
	}

	hashesInCertificate := make([][]byte, 0)

	//	collect hash from pre-prepare message
	proposalHash := messages.ExtractProposalHash(certificate.GetProposalMessage())
	hashesInCertificate = append(hashesInCertificate, proposalHash)

	//	collect hashes from prepare messages
	for _, msg := range certificate.GetPrepareMessages() {
		proposalHash := messages.ExtractPrepareHash(msg)

		hashesInCertificate = append(hashesInCertificate, proposalHash)
	}

	//	verify all hashes match the proposal
	for _, hash := range hashesInCertificate {
		if !i.backend.IsValidProposalHash(proposal, hash) {
			return false
		}
	}

	return true
}

// runStates is the main loop which performs state transitions
func (i *IBFT) runStates(ctx context.Context) {
	var timeout error

	for {
		switch i.state.getStateName() {
		case newRound:
			timeout = i.runNewRound(ctx)
		case prepare:
			timeout = i.runPrepare(ctx)
		case commit:
			timeout = i.runCommit(ctx)
		case fin:
			i.runFin(ctx)

			return
		}

		if timeout != nil {
			// Timeout received
			return
		}
	}
}

// runNewRound runs the New Round IBFT state
func (i *IBFT) runNewRound(ctx context.Context) error {
	startTime := time.Now()
	i.log.Debug("enter: new round state")
	defer i.log.Debug("exit: new round state")
	defer i.setMeasurementTime(ctx, i.newRoundHistogram, startTime)

	var (
		// Grab the current view
		view = i.state.getView()

		// Subscribe for PREPREPARE messages
		sub = i.subscribe(
			messages.SubscriptionDetails{
				MessageType: proto.MessageType_PREPREPARE,
				View:        view,
			},
		)
	)

	// The subscription is not needed anymore after
	// this state is done executing
	defer i.messages.Unsubscribe(sub.ID)

	for {
		select {
		case <-ctx.Done():
			// Stop signal received, exit
			return errTimeoutExpired
		case <-sub.SubCh:
			// SubscriptionDetails conditions have been met,
			// grab the proposal messages
			proposalMessage := i.handlePrePrepare(view)
			if proposalMessage == nil {
				continue
			}

			// Multicast the PREPARE message
			i.state.setProposalMessage(proposalMessage)
			i.sendPrepareMessage(view)

			i.log.Debug("prepare message multicasted")

			// Move to the prepare state
			i.state.changeState(prepare)

			return nil
		}
	}
}

// validateProposalCommon does common validations for each proposal, no
// matter the round
func (i *IBFT) validateProposalCommon(msg *proto.IbftMessage, view *proto.View) bool {
	var (
		height = view.GetHeight()
		round  = view.GetRound()

		proposal     = messages.ExtractProposal(msg)
		proposalHash = messages.ExtractProposalHash(msg)
	)

	//	round matches
	if proposal.GetRound() != view.GetRound() {
		return false
	}

	//	is proposer
	if !i.backend.IsProposer(msg.GetFrom(), height, round) {
		return false
	}

	//	hash matches keccak(proposal)
	if !i.backend.IsValidProposalHash(proposal, proposalHash) {
		return false
	}

	//	is valid proposal
	return i.backend.IsValidProposal(proposal.GetRawProposal())
}

// validateProposal0 validates the proposal for round 0
func (i *IBFT) validateProposal0(msg *proto.IbftMessage, view *proto.View) bool {
	var (
		height = view.GetHeight()
		round  = view.GetRound()
	)

	//	proposal must be for round 0
	if msg.GetView().GetRound() != 0 {
		return false
	}

	// Make sure common proposal validations pass
	if !i.validateProposalCommon(msg, view) {
		return false
	}

	// Make sure the current node is not the proposer for this round
	if i.backend.IsProposer(i.backend.ID(), height, round) {
		return false
	}

	return true
}

// validateProposal validates a proposal for round > 0
func (i *IBFT) validateProposal(msg *proto.IbftMessage, view *proto.View) bool {
	var (
		height = view.GetHeight()
		round  = view.GetRound()

		proposal = messages.ExtractProposal(msg)
		rcc      = messages.ExtractRoundChangeCertificate(msg)
	)

	// Make sure common proposal validations pass
	if !i.validateProposalCommon(msg, view) {
		return false
	}

	// Make sure there is a certificate
	if rcc == nil {
		return false
	}

	// Make sure all the messages have the unique sender
	if !messages.HasUniqueSenders(rcc.GetRoundChangeMessages()) {
		return false
	}

	// Make sure there are Quorum RCC
	if !i.hasQuorumByMsgType(rcc.GetRoundChangeMessages(), proto.MessageType_ROUND_CHANGE) {
		return false
	}

	// Make sure the current node is not the proposer for this round
	if i.backend.IsProposer(i.backend.ID(), height, round) {
		return false
	}

	// Make sure all messages in the RCC are valid Round Change messages
	for _, rc := range rcc.GetRoundChangeMessages() {
		// Make sure the message is a Round Change message
		if rc.GetType() != proto.MessageType_ROUND_CHANGE {
			return false
		}

		// Height of the message matches height of the proposal
		if rc.GetView().GetHeight() != height {
			return false
		}

		// Round of the message matches round of the proposal
		if rc.GetView().GetRound() != round {
			return false
		}

		// Sender of RCC is valid
		if !i.backend.IsValidValidator(rc) {
			return false
		}
	}

	// Extract possible rounds and their corresponding
	// block hashes
	type roundHashTuple struct {
		round uint64
		hash  []byte
	}

	roundsAndPreparedBlockHashes := make([]roundHashTuple, 0)

	for _, rcMessage := range rcc.GetRoundChangeMessages() {
		cert := messages.ExtractLatestPC(rcMessage)

		// Check if there is a certificate, and if it's a valid PC
		if cert != nil && i.validPC(cert, msg.GetView().GetRound(), height) {
			hash := messages.ExtractProposalHash(cert.GetProposalMessage())

			roundsAndPreparedBlockHashes = append(roundsAndPreparedBlockHashes, roundHashTuple{
				round: cert.GetProposalMessage().GetView().GetRound(),
				hash:  hash,
			})
		}
	}

	if len(roundsAndPreparedBlockHashes) == 0 {
		return true
	}

	// Find the max round
	var (
		maxRound     uint64
		expectedHash []byte
	)

	for _, tuple := range roundsAndPreparedBlockHashes {
		if tuple.round >= maxRound {
			maxRound = tuple.round
			expectedHash = tuple.hash
		}
	}

	// Make sure hash of (EB, maxR) matches expected hash
	return i.backend.IsValidProposalHash(
		&proto.Proposal{
			RawProposal: proposal.GetRawProposal(),
			Round:       maxRound,
		},
		expectedHash,
	)
}

// handlePrePrepare parses the received proposal and performs
// a transition to PREPARE state, if the proposal is valid
func (i *IBFT) handlePrePrepare(view *proto.View) *proto.IbftMessage {
	isValidPrePrepare := func(message *proto.IbftMessage) bool {
		if view.GetRound() == 0 {
			//	proposal must be for round 0
			return i.validateProposal0(message, view)
		}

		return i.validateProposal(message, view)
	}

	msgs := i.messages.GetValidMessages(
		view,
		proto.MessageType_PREPREPARE,
		isValidPrePrepare,
	)

	if len(msgs) < 1 {
		return nil
	}

	return msgs[0]
}

// runPrepare runs the Prepare IBFT state
func (i *IBFT) runPrepare(ctx context.Context) error {
	startTime := time.Now()
	i.log.Debug("enter: prepare state")
	defer i.log.Debug("exit: prepare state")
	defer i.setMeasurementTime(ctx, i.prepareHistogram, startTime)

	var (
		// Grab the current view
		view = i.state.getView()

		// Subscribe to PREPARE messages
		sub = i.subscribe(
			messages.SubscriptionDetails{
				MessageType: proto.MessageType_PREPARE,
				View:        view,
			},
		)
	)

	// The subscription is not needed anymore after
	// this state is done executing
	defer i.messages.Unsubscribe(sub.ID)

	for {
		select {
		case <-ctx.Done():
			// Stop signal received, exit
			return errTimeoutExpired
		case <-sub.SubCh:
			if !i.handlePrepare(view) {
				//	quorum of valid prepare messages not received, retry
				continue
			}

			return nil
		}
	}
}

// handlePrepare parses available prepare messages and performs
// a transition to COMMIT state, if quorum was reached
func (i *IBFT) handlePrepare(view *proto.View) bool {
	isValidPrepare := func(message *proto.IbftMessage) bool {
		// Verify that the proposal hash is valid
		return i.backend.IsValidProposalHash(
			i.state.getProposal(),
			messages.ExtractPrepareHash(message),
		)
	}

	prepareMessages := i.messages.GetValidMessages(
		view,
		proto.MessageType_PREPARE,
		isValidPrepare,
	)

	if !i.hasQuorumByMsgType(prepareMessages, proto.MessageType_PREPARE) {
		//	quorum not reached, keep polling
		return false
	}

	// Multicast the COMMIT message
	i.sendCommitMessage(view)

	i.log.Debug("commit message multicasted")

	i.state.finalizePrepare(
		&proto.PreparedCertificate{
			ProposalMessage: i.state.getProposalMessage(),
			PrepareMessages: prepareMessages,
		},
		i.state.getProposal(),
	)

	return true
}

// runCommit runs the Commit IBFT state
func (i *IBFT) runCommit(ctx context.Context) error {
	startTime := time.Now()
	i.log.Debug("enter: commit state")
	defer i.log.Debug("exit: commit state")
	defer i.setMeasurementTime(ctx, i.commitHistogram, startTime)

	var (
		// Grab the current view
		view = i.state.getView()

		// Subscribe to COMMIT messages
		sub = i.subscribe(
			messages.SubscriptionDetails{
				MessageType: proto.MessageType_COMMIT,
				View:        view,
			},
		)
	)

	// The subscription is not needed anymore after
	// this state is done executing
	defer i.messages.Unsubscribe(sub.ID)

	for {
		select {
		case <-ctx.Done():
			// Stop signal received, exit
			return errTimeoutExpired
		case <-sub.SubCh:
			if !i.handleCommit(view) {
				//	quorum not reached, retry
				continue
			}

			return nil
		}
	}
}

// handleCommit parses available commit messages and performs
// a transition to FIN state, if quorum was reached
func (i *IBFT) handleCommit(view *proto.View) bool {
	isValidCommit := func(message *proto.IbftMessage) bool {
		var (
			proposalHash  = messages.ExtractCommitHash(message)
			committedSeal = messages.ExtractCommittedSeal(message)
		)
		//	Verify that the proposal hash is valid
		if !i.backend.IsValidProposalHash(i.state.getProposal(), proposalHash) {
			return false
		}

		//	Verify that the committed seal is valid
		return i.backend.IsValidCommittedSeal(proposalHash, committedSeal)
	}

	commitMessages := i.messages.GetValidMessages(view, proto.MessageType_COMMIT, isValidCommit)
	if !i.hasQuorumByMsgType(commitMessages, proto.MessageType_COMMIT) {
		//	quorum not reached, keep polling
		return false
	}

	commitSeals, err := messages.ExtractCommittedSeals(commitMessages)
	if err != nil {
		// safe check
		i.log.Error("failed to extract committed seals from commit messages", "error", err)

		return false
	}

	// Set the committed seals
	i.state.setCommittedSeals(commitSeals)

	//	Move to the fin state
	i.state.changeState(fin)

	return true
}

// runFin runs the fin state (block insertion)
func (i *IBFT) runFin(ctx context.Context) {
	i.log.Debug("enter: fin state")
	defer i.log.Debug("exit: fin state")

	i.signalRoundDone(ctx)
}

// insertBlock inserts the block
func (i *IBFT) insertBlock() {
	// Insert the block to the node's underlying
	// blockchain layer
	i.backend.InsertProposal(
		&proto.Proposal{
			RawProposal: i.state.getRawDataFromProposal(),
			Round:       i.state.getRound(),
		},
		i.state.getCommittedSeals(),
	)

	// Remove stale messages
	i.messages.PruneByHeight(i.state.getHeight())
}

// moveToNewRound moves the state to the new round
func (i *IBFT) moveToNewRound(round uint64) {
	i.state.setView(&proto.View{
		Height: i.state.getHeight(),
		Round:  round,
	})

	i.state.setRoundStarted(false)
	i.state.setProposalMessage(nil)
	i.state.changeState(newRound)
}

func (i *IBFT) buildProposal(ctx context.Context, view *proto.View) *proto.IbftMessage {
	var (
		height = view.GetHeight()
		round  = view.GetRound()
	)

	if round == 0 {
		rawProposal := i.backend.BuildProposal(
			&proto.View{
				Height: height,
				Round:  round,
			})

		return i.backend.BuildPrePrepareMessage(
			rawProposal,
			nil,
			&proto.View{
				Height: height,
				Round:  round,
			},
		)
	}

	//	round > 0 -> needs RCC
	rcc := i.waitForRCC(ctx, height, round)
	if rcc == nil {
		// Timeout occurred
		i.log.Error("no valid RCC received, timeout expired")
		return nil
	}

	//	check the messages for any previous proposal (if they have any, it's the same proposal)
	var (
		previousProposal []byte
		maxRound         uint64
	)

	// take previous proposal among the round change messages for the highest round
	for _, msg := range rcc.GetRoundChangeMessages() {
		latestPC := messages.ExtractLatestPC(msg)
		if latestPC == nil {
			continue
		}

		proposal := messages.ExtractProposal(latestPC.GetProposalMessage())
		round := proposal.GetRound()

		// skip if message's round is equals to/less than maxRound
		if previousProposal != nil && round <= maxRound {
			continue
		}

		lastPB := messages.ExtractLastPreparedProposal(msg)
		if lastPB == nil {
			continue
		}

		previousProposal = lastPB.GetRawProposal()
		maxRound = round
	}

	proposal := previousProposal
	if proposal == nil {
		//	build new proposal
		proposal = i.backend.BuildProposal(
			&proto.View{
				Height: height,
				Round:  round,
			})
	}

	return i.backend.BuildPrePrepareMessage(
		proposal,
		rcc,
		&proto.View{
			Height: height,
			Round:  round,
		},
	)
}

// acceptProposal accepts the proposal and moves the state
func (i *IBFT) acceptProposal(proposalMessage *proto.IbftMessage) {
	//	accept newly proposed block and move to PREPARE state
	i.state.setProposalMessage(proposalMessage)
	i.state.changeState(prepare)
}

// AddMessage adds a new message to the IBFT message system
func (i *IBFT) AddMessage(message *proto.IbftMessage) {
	// Make sure the message is present
	if message == nil {
		return
	}

	// Check if the message should even be considered
	if i.isAcceptableMessage(message) {
		i.messages.AddMessage(message)

		// Signal event if the quorum is reached. Since the subscriptions refer to the state height,
		// no need to call this if the message height is not equal to the state height
		if message.GetView().GetHeight() == i.state.getHeight() {
			msgs := i.messages.GetValidMessages(
				message.GetView(),
				message.GetType(),
				func(_ *proto.IbftMessage) bool { return true })
			if i.hasQuorumByMsgType(msgs, message.GetType()) {
				i.messages.SignalEvent(message.GetType(), message.GetView())
			}
		}
	}
}

// isAcceptableMessage checks if the message can even be accepted
func (i *IBFT) isAcceptableMessage(message *proto.IbftMessage) bool {
	// Invalid messages are discarded
	if message.GetView() == nil {
		return false
	}

	// Make sure the message is in accordance with
	// the current state height, or greater
	if i.state.getHeight() > message.GetView().GetHeight() {
		return false
	}

	//	Make sure the message sender is ok
	if !i.backend.IsValidValidator(message) {
		return false
	}

	// Make sure if the heights are the same, the message round is >= the current state round
	if i.state.getHeight() == message.GetView().GetHeight() {
		return message.GetView().GetRound() >= i.state.getRound()
	}

	return true
}

// ExtendRoundTimeout extends each round's timer by the specified amount.
func (i *IBFT) ExtendRoundTimeout(amount time.Duration) {
	i.additionalTimeout = amount
}

// validPC verifies that the prepared certificate is valid
func (i *IBFT) validPC(
	certificate *proto.PreparedCertificate,
	roundLimit,
	height uint64,
) bool {
	if certificate == nil {
		// PCs that are not set are valid by default
		return true
	}

	// Make sure that either both the proposal message and the prepare messages are set together
	if certificate.GetProposalMessage() == nil || certificate.PrepareMessages == nil {
		return false
	}

	allMessages := append(
		[]*proto.IbftMessage{certificate.GetProposalMessage()},
		certificate.GetPrepareMessages()...,
	)

	// Make sure there are at least Quorum (PP + P) messages
	// hasQuorum directly since the messages are of different types
	if !i.validatorManager.HasQuorum(convertMessageToAddressSet(allMessages)) {
		return false
	}

	// Make sure the proposal message is a Preprepare message
	if certificate.GetProposalMessage().GetType() != proto.MessageType_PREPREPARE {
		return false
	}

	// Make sure all messages in the PC are Prepare messages
	for _, message := range certificate.GetPrepareMessages() {
		if message.GetType() != proto.MessageType_PREPARE {
			return false
		}
	}

	// Make sure the round, height and proposal hashes match and the senders are unique
	if !messages.AreValidPCMessages(allMessages, height, roundLimit) {
		return false
	}

	// Make sure the proposal message is sent by the proposer
	// for the round
	proposal := certificate.GetProposalMessage()
	if !i.backend.IsProposer(proposal.GetFrom(), proposal.GetView().GetHeight(), proposal.GetView().GetRound()) {
		return false
	}

	// Make sure that the proposal sender is valid
	if !i.backend.IsValidValidator(proposal) {
		return false
	}

	// Make sure the Prepare messages are validators, apart from the proposer
	for _, message := range certificate.GetPrepareMessages() {
		// Make sure the sender is part of the validator set
		if !i.backend.IsValidValidator(message) {
			return false
		}

		// Make sure the current node is not the proposer
		if i.backend.IsProposer(message.GetFrom(), message.GetView().GetHeight(), message.GetView().GetRound()) {
			return false
		}
	}

	return true
}

// sendPreprepareMessage sends out the preprepare message
func (i *IBFT) sendPreprepareMessage(message *proto.IbftMessage) {
	i.transport.Multicast(message)
}

// sendRoundChangeMessage sends out the round change message
func (i *IBFT) sendRoundChangeMessage(height, newRound uint64) {
	i.transport.Multicast(
		i.backend.BuildRoundChangeMessage(
			i.state.getLatestPreparedProposal(),
			i.state.getLatestPC(),
			&proto.View{
				Height: height,
				Round:  newRound,
			},
		),
	)
}

// sendPrepareMessage sends out the prepare message
func (i *IBFT) sendPrepareMessage(view *proto.View) {
	i.transport.Multicast(
		i.backend.BuildPrepareMessage(
			i.state.getProposalHash(),
			view,
		),
	)
}

// sendCommitMessage sends out the commit message
func (i *IBFT) sendCommitMessage(view *proto.View) {
	i.transport.Multicast(
		i.backend.BuildCommitMessage(
			i.state.getProposalHash(),
			view,
		),
	)
}

// hasQuorumByMsgType provides information on whether messages of specific types have reached the quorum
func (i *IBFT) hasQuorumByMsgType(msgs []*proto.IbftMessage, msgType proto.MessageType) bool {
	switch msgType {
	case proto.MessageType_PREPREPARE:
		return len(msgs) >= 1
	case proto.MessageType_PREPARE:
		return i.validatorManager.HasPrepareQuorum(i.state.getStateName(), i.state.getProposalMessage(), msgs)
	case proto.MessageType_ROUND_CHANGE, proto.MessageType_COMMIT:
		return i.validatorManager.HasQuorum(convertMessageToAddressSet(msgs))
	default:
		return false
	}
}

func (i *IBFT) subscribe(details messages.SubscriptionDetails) *messages.Subscription {
	subscription := i.messages.Subscribe(details)
	msgs := i.messages.GetValidMessages(
		details.View,
		details.MessageType,
		func(_ *proto.IbftMessage) bool { return true })
	// Check if any condition is already met
	if i.hasQuorumByMsgType(msgs, details.MessageType) {
		i.messages.SignalEvent(details.MessageType, details.View)
	}

	return subscription
}

// getRoundTimeout creates a round timeout based on the base timeout and the current round.
// Exponentially increases timeout depending on the round number.
// For instance:
//   - round 1: 1 sec
//   - round 2: 2 sec
//   - round 3: 4 sec
//   - round 4: 8 sec
func getRoundTimeout(baseRoundTimeout, additionalTimeout time.Duration, round uint64) time.Duration {
	var (
		duration     = int(baseRoundTimeout)
		roundFactor  = int(math.Pow(roundFactorBase, float64(round)))
		roundTimeout = time.Duration(duration * roundFactor)
	)

	return roundTimeout + additionalTimeout
}
