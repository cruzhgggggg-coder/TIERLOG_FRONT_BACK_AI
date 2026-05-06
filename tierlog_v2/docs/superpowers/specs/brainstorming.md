# Brainstorming: Cyber-Academic Mastery & Performance Sovereignty

## Purpose
To finalize the "Cyber-Academic" visual upgrade for TierLog and implement performance optimizations (List Virtualization, Canvas Profiling) to ensure a "Master Mind" grade experience.

## Finalized Decisions

### 1. Visual Direction: Minimalist Lab-Tech
- **Theme**: Void Black base with high-precision Synaptic Cyan accents.
- **Feedback Coding**:
    - **HOC**: Logic Magenta (#ff00ea) - represents structural/logical flow.
    - **LOC**: Synaptic Cyan (#00f2ff) - represents technical/surface details.
- **Interactions**:
    - **Reactive Canvas**: The `NeuralNetwork` background will "cluster" nodes around the active feedback item.
    - **Glassmorphism**: 3-column layout will use varying blur levels to create depth.

### 2. Performance Architecture
- **Virtualization**: Implement `RecycleScroller` from `vue-virtual-scroller` for the Feedback Repository (Center Column) to handle 100+ feedback items without lag.
- **Canvas Batching**: Refactor `NeuralNetwork.vue` to use a single `requestAnimationFrame` loop with optimized node calculations.

### 3. Actionable Intelligence (Neural Actions)
- **Decision**: Collaborative Sub-Chat.
- **Feature**: When "Ask AI to Fix" is clicked, it highlights the feedback item in the Center Column and opens a focused sub-context in the Oracle (Right Column).
- **Neural Patching**: AI responses will include a `[PATCH]` block that can be visually compared.

## Technical Implementation Notes
- **Dependencies**: `npm install vue-virtual-scroller`.
- **CSS**: Upgrade `_tokens.css` with dynamic CSS variables for "active" states.

---
*Document locked for implementation planning.*
