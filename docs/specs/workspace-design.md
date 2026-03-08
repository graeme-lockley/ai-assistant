# workspace-design.md
Unified AI Agent Workspace Architecture

---

# 1. Purpose

This document defines the architecture of an AI agent workspace inspired by OpenClaw-style agents.

The workspace is the persistent cognitive substrate of the agent. It stores identity,
memory, capabilities, and interaction history in a structured manner. The design is
intended to support the agent's interaction with the environment **through the passing
of time**: as sessions accumulate, memory is consolidated, and the agent learns about
the user and its role, the agent should **become more relevant**, not less. These
documents are the basis for continuing development of the workspace and the agent.

Design goals:

- Stable agent identity
- Compressed long-term memory
- Explicit capabilities (skills and tools)
- Minimal prompt context (as small as possible while correct; no loss of fidelity from squeezing into token budgets)
- Scalable cognition
- Predictable behaviour over long periods

The design mirrors elements of human cognition.

| Human Analogue | Agent Equivalent |
|----------------|-----------------|
| Personality | SOUL.md |
| Identity | IDENTITY.md |
| Current role | AGENT.md |
| Understanding of the user | USER.md |
| Long-term memory | MEMORY.md |
| Skills | skills/ |
| Tools | tools/ |
| Experiences | logs/ |
| Sleep consolidation | memory/ |

---

# 2. High Level Architecture

The workspace contains five conceptual layers.

Operational Self
- AGENT.md
- IDENTITY.md
- SOUL.md

User Context
- USER.md
- MEMORY.md

Capabilities
- SKILLS.md
- TOOLS.md
- skills/
- tools/

Operational State
- TASKS.md

Historical Substrate
- logs/
- memory/

Context Assembly
- context loader (see [context-loader-spec](context-loader-spec.md))

---

# 3. Directory Structure

The design document (this file) lives under `docs/specs/`. The workspace on disk:

workspace/

AGENT.md  
IDENTITY.md  
SOUL.md  
USER.md  
MEMORY.md  

SKILLS.md  
TOOLS.md  

TASKS.md  

WORKSPACE.md  

logs/  
memory/  
skills/  
tools/  
context/  

The `context/` directory (loader configuration, indexes, routing, cache) is defined in [context-loader-spec](context-loader-spec.md). The role of WORKSPACE.md is also specified there.

---

# 4. Core Files

## AGENT.md

Defines the operational *what*: how the agent should operate in its role as an assistant.

It answers: what is the agent instructed to do in this role?

Typical content:

- mission
- operating principles
- constraints
- reasoning preferences
- decision guidelines

---

## IDENTITY.md

Defines the agent's stable self‑model—*who* the agent is.

Identity evolves slowly as the agent learns about its role; changes are under the agent's control through conversation with the user.

Content includes:

- name
- purpose
- domain
- strengths
- boundaries

---

## SOUL.md

Defines the agent's beliefs, temperament, and communication style.

SOUL governs values and tone rather than operational instructions. In conflicts with other files, SOUL has the highest gravity (see §9).

Typical content:

- tone
- values
- philosophy
- conversational style

---

## USER.md

Represents what the agent knows about the user: durable, stable facts.

Typical sections

Identity  
Preferences  
Projects  
Interests  
Constraints

---

## MEMORY.md

Contains distilled long‑term knowledge from user–agent interactions.

MEMORY.md is a high‑signal synthesis, not a transcript. It is fed by the consolidation pipeline that runs over `memory/` (see §6). Agent, Soul, and Identity are what the agent knows about itself; User and Memory are what it knows about the user and past interactions.

Possible sections

Facts  
Preferences  
Active Threads  
Beliefs  
Open Questions

---

## TASKS.md

Tracks open commitments and ongoing work. The agent updates TASKS.md when necessary.

---

# 5. Capability System

Capabilities are split into skills and tools.

- **Skills** are added into the system through markdown files. Each skill file explains the purpose of the skill and the different actions or elements that make it up. Skills live in `skills/` and are described in `SKILLS.md`.
- **Tools** are built into the agent (implemented in code). They are registered and described in the workspace; the executable behaviour is part of the agent runtime.

SKILLS.md and TOOLS.md are summaries with a short description of purpose for each capability; see [context-loader-spec](context-loader-spec.md) for how they are used in context assembly.

---

# 6. Historical Storage

## logs/

Raw session transcripts. The system appends all content to the session log as the conversation proceeds.

Logs are:

- append‑only
- never loaded into prompts (unless explicitly stated otherwise in the prompt)
- used only for summarisation and offline consolidation

---

## memory/

Stores hierarchical summaries produced by the consolidation pipeline.

Flow: `memory/` (daily → weekly → monthly → annual summaries) is distilled into **MEMORY.md**. Distillation is performed by separate jobs that run at night (e.g. via cron). A script will use prompts to perform the distillation. This "sleep consolidation" keeps long-term memory compressed and high-signal.

---

# 7. Context Loader

The Context Loader is part of the workspace system. It assembles the minimum working context required for each request.

Responsibilities: classify the request; retrieve relevant workspace fragments; rank them; compose the final prompt context. Selection and ranking behaviour are specified in [context-loader-spec](context-loader-spec.md).

---

# 8. Context Rings

Ring 1 — Core Self

SOUL.md  
AGENT.md  
IDENTITY.md  

Ring 2 — Supporting Context

selected USER fragments (see [context-loader-spec](context-loader-spec.md))  
selected MEMORY fragments  
TASKS.md items  

Ring 3 — Capability Context

relevant skills  
relevant tools

---

# 9. Minimal System Prompt

You are an AI agent operating inside a structured workspace.

The workspace defines your identity, behaviour, memory, skills, and tools.

Follow the workspace files in this priority order (earlier has greater gravity in case of conflict):

1. SOUL.md
2. AGENT.md
3. IDENTITY.md
4. USER.md
5. MEMORY.md

Rules

SOUL.md defines your beliefs, tone, and values.  
AGENT.md defines how you operate in your role.  
IDENTITY.md defines who you are.  
USER.md describes the user.  
MEMORY.md contains distilled long‑term knowledge.

Skills and tools may be loaded when required.

Never load raw logs (unless explicitly stated otherwise in the prompt).

Use structured reasoning when solving problems.

Update TASKS.md when necessary.

Remain consistent with your identity and memory.

---

# 10. Summary

This workspace architecture enables persistent AI agents that remain stable and useful over time.

It ensures: stable identity; compressed long‑term memory; explicit capabilities; minimal prompt size; scalable cognition.

**Evolution over time.** The architecture is built so that development can continue and the agent can become more relevant as time passes: new skills are added via markdown; USER.md and MEMORY.md grow from interaction and consolidation; IDENTITY.md may evolve slowly through conversation; logs and the memory hierarchy feed nightly distillation into MEMORY.md. The context loader (see [context-loader-spec](context-loader-spec.md)) keeps each turn focused on the right subset of this growing workspace, so the agent scales without losing coherence or relevance.

This document and [context-loader-spec](context-loader-spec.md) should be kept in sync: the Context Loader implements the context assembly described here, and the spec defines retrieval, ranking, selection, and the roles of `context/` and WORKSPACE.md.
