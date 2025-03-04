package ibft

import (
	"context"

	"github.com/NilFoundation/nil/nil/internal/telemetry"
	"github.com/NilFoundation/nil/nil/internal/telemetry/telattr"
	"github.com/NilFoundation/nil/nil/internal/types"
	"go.opentelemetry.io/otel/metric"
)

type MetricsHandler struct {
	option metric.MeasurementOption

	buildProposalMeasurer  *telemetry.Measurer
	insertProposalMeasurer *telemetry.Measurer
	sequenceMeasurer       *telemetry.Measurer

	height           telemetry.Histogram
	round            telemetry.Histogram
	validatorsCount  telemetry.Histogram
	sentMessages     telemetry.Counter
	receivedMessages telemetry.Counter
}

func NewMetricsHandler(name string, shardId types.ShardId) (*MetricsHandler, error) {
	meter := telemetry.NewMeter(name)
	buildProposalMeasurer, err := telemetry.NewMeasurer(
		meter, "build_proposal", telattr.ShardId(shardId),
	)
	if err != nil {
		return nil, err
	}

	insertProposalMeasurer, err := telemetry.NewMeasurer(
		meter, "insert_proposal", telattr.ShardId(shardId),
	)
	if err != nil {
		return nil, err
	}

	sequenceMeasurer, err := telemetry.NewMeasurer(
		meter, "sequence", telattr.ShardId(shardId),
	)
	if err != nil {
		return nil, err
	}

	handler := &MetricsHandler{
		buildProposalMeasurer:  buildProposalMeasurer,
		insertProposalMeasurer: insertProposalMeasurer,
		sequenceMeasurer:       sequenceMeasurer,
		option:                 telattr.With(telattr.ShardId(shardId)),
	}

	if err := handler.initMetrics(meter); err != nil {
		return nil, err
	}

	return handler, nil
}

func (mh *MetricsHandler) initMetrics(meter metric.Meter) error {
	var err error

	if mh.validatorsCount, err = meter.Int64Histogram("validators_count"); err != nil {
		return err
	}

	if mh.height, err = meter.Int64Histogram("height"); err != nil {
		return err
	}

	if mh.round, err = meter.Int64Histogram("round"); err != nil {
		return err
	}

	if mh.sentMessages, err = meter.Int64Counter("sent_messages"); err != nil {
		return err
	}

	if mh.receivedMessages, err = meter.Int64Counter("received_messages"); err != nil {
		return err
	}

	return nil
}

func (mh *MetricsHandler) StartBuildProposalMeasurement(ctx context.Context, round uint64) {
	mh.round.Record(ctx, int64(round), mh.option)
	mh.buildProposalMeasurer.Restart()
}

func (mh *MetricsHandler) EndBuildProposalMeasurement(ctx context.Context) {
	mh.buildProposalMeasurer.Measure(ctx)
}

func (mh *MetricsHandler) StartInsertProposalMeasurement(ctx context.Context, round uint64) {
	mh.round.Record(ctx, int64(round), mh.option)
	mh.insertProposalMeasurer.Restart()
}

func (mh *MetricsHandler) EndInsertProposalMeasurement(ctx context.Context, height, round uint64) {
	mh.insertProposalMeasurer.Measure(ctx)
}

func (mh *MetricsHandler) StartSequenceMeasurement(ctx context.Context, height uint64) {
	mh.height.Record(ctx, int64(height), mh.option)
	mh.round.Record(ctx, 0, mh.option)
	mh.sequenceMeasurer.Restart()
}

func (mh *MetricsHandler) EndSequenceMeasurement(ctx context.Context) {
	mh.sequenceMeasurer.Measure(ctx)
}

func (mh *MetricsHandler) SetValidatorsCount(ctx context.Context, count int) {
	mh.validatorsCount.Record(ctx, int64(count), mh.option)
}

func (mh *MetricsHandler) IncSentMessages(ctx context.Context, t string) {
	mh.sentMessages.Add(ctx, 1, mh.option, telattr.With(telattr.Type(t)))
}

func (mh *MetricsHandler) IncReceivedMessages(ctx context.Context, t string) {
	mh.receivedMessages.Add(ctx, 1, mh.option, telattr.With(telattr.Type(t)))
}
