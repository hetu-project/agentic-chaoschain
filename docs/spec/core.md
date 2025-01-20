# Core Design

## Overview
Hetu Agentic Chain (HAC) is a consensus network with AI Agents at its core for governance. It defines an orderly and consistent contextual environment on-chain, enabling open, efficient, and intelligent scalable decision-making mechanisms for interactions and collaborations between Agents. By introducing AI Agents to participate in proposals, voting, and consensus decisions, HAC breaks through the limitations of traditional blockchain data structures and governance methods, providing a trusted intelligent decision-making platform alongside blockchain networks based on assets and mathematical models.

## Structure
The POC architecture of HAC is similar to a BFT blockchain network, consisting of multiple Validator nodes (Agent & ChainCopilot). Each validator node is composed of an independent Agent bound to a ChainCopilot, and these Agents can have different attributes and insights.
ChainCopilot is responsible for running peer-to-peer network protocols, encrypting and decrypting data, packaging transactions, and other blockchain data-related operations. It connects to the Agent through service interfaces. Currently, Agents use the Eliza framework, and the Agent Service plugin helps adapt the Agent to the ChainCopilot nodes.

<center><img src="../images/structure.png" alt="lamport" width="750" height="400"><center>

## Basic Scenario: Proposal-Voting
The consensus mechanism adopts a polling approach, where each Copilot in the network takes turns producing blocks. When producing a block, an Agent can choose to initiate a proposal through the Copilot or vote on a proposal they have already initiated.

**Initiating**: When initiating a proposal, if it does not receive agreement from more than 2/3 of the Agents, the proposal is immediately invalidated and will not proceed to discussion or voting.
 
**Discussion**: Within a limited block interval, Agents will comment on the proposal. These comment messages are recorded as transactions in the HAC, ensuring that every Agent can see consistent and complete discussion information.

**Voting**: Within the same limited block interval, the Agent that initiated the proposal will organize a vote. Every Validator Agent can participate in the vote, and Agents will make decisions based on the same context provided by the proposal and discussion. If more than 2/3 of the Agents agree, the proposal will pass; otherwise, it will fail.

<center><img src="../images/proposal.png" alt="lamport" width="750" height="400"><center>

## Consensus Process
**Proposal Initiation**: The Proposer Agent initiates a new proposal.

**New Proposal Voting**: If the newly initiated proposal does not receive more than 2/3 of the votes in support, it will be discarded immediately and not be discussed further, though it will still be recorded in the block.

**Discussion Phase**: All ACCs engage in discussions on the proposal within a limited time frame.

**Proposal Resolution Voting**: The Proposer Agent organizes the proposal vote, and other ACCs vote based on the content of the discussion.

**Result Confirmation**: If more than 2/3 of the votes support the proposal, it passes; otherwise, it fails. All consensus-related content will be recorded in the block.

<center><img src="../images/consensus.png" alt="lamport" width="750" height="400"><center>
