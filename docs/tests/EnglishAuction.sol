// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import "@nilfoundation/smart-contracts/contracts/Nil.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title EnglishAuction
 * @author =nil; Foundation
 * @notice This contract implements an auction where contracts can place bids
 * @notice and the contract owner decides when to start and end the auction.
 */
contract EnglishAuction is Ownable {
    event Start();
    event Bid(address indexed sender, uint256 amount);
    event Withdraw(address indexed bidder, uint256 amount);
    event End(address winner, uint256 amount);

    //startContractProperties
    /**
     * @notice These properties store the address of the NFT contract
     * and check whether the auction is still going.
     */
    address private nft;
    bool public isOngoing;

    /**
     * @notice These properties store information about all bids as well as
     * the current highest bid and bidder.
     */
    address public highestBidder;
    uint256 public highestBid;
    mapping(address => uint256) public bids;

    /**
     * @notice The constructor stores the address of the NFT contract
     * and accepts the initial bid.
     * @param _nft The address of the NFT contract.
     */
    constructor(address _nft) payable Ownable(msg.sender) {
        nft = _nft;
        isOngoing = false;
        highestBid = msg.value;
    }

    //endContractProperties

    /**
     * @notice This function starts the auction and sends a transaction
     * for minting the NFT.
     */
    function start() public onlyOwner {
        require(!isOngoing, "the auction has already started");

        Nil.asyncCall(
            nft,
            address(this),
            address(this),
            0,
            Nil.FORWARD_REMAINING,
            0,
            abi.encodeWithSignature("mintNFT()")
        );

        isOngoing = true;

        emit Start();
    }

    //startAuctionLogic
    /**
     * @notice The function submits a bid for the auction.
     */
    function bid() public payable {
        require(isOngoing, "the auction has not started");
        require(
            msg.value > highestBid,
            "the bid does not exceed the current highest bid"
        );

        if (highestBidder != address(0)) {
            bids[highestBidder] += highestBid;
        }

        highestBidder = msg.sender;
        highestBid = msg.value;

        emit Bid(msg.sender, msg.value);
    }

    /**
     * @notice This function exists so a bidder can withdraw their funds
     * if they change their mind.
     */
    function withdraw() public {
        uint256 bal = bids[msg.sender];
        bids[msg.sender] = 0;

        Nil.asyncCall(msg.sender, address(this), bal, "");

        emit Withdraw(msg.sender, bal);
    }

    /**
     * @notice This function ends the auction and requests the NFT contract
     * to provide the NFT to the winner.
     */
    function end() public onlyOwner {
        require(isOngoing, "the auction has not started");

        isOngoing = false;

        Nil.asyncCall(
            nft,
            address(this),
            address(this),
            0,
            Nil.FORWARD_REMAINING,
            0,
            abi.encodeWithSignature("sendNFT(address)", highestBidder)
        );

        emit End(highestBidder, highestBid);
    }
    //endAuctionLogic
}
