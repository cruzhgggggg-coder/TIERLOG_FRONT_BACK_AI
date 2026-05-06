# Implementation Plan: Cyber-Academic Neural Interface Upgrade

Refactoring TierLog's consultation workspace into a high-performance, actionable neural interface.

## User Review Required

> [!IMPORTANT]
> This upgrade introduces `vue-virtual-scroller` which requires a new dependency.
> The visual palette for HOC/LOC is being changed to Logic Magenta and Synaptic Cyan for better alignment with the CADS vision.

## Proposed Changes

### 1. Style & Tokens
Update `_tokens.css` to include the new color logic and glassmorphism levels.

#### [MODIFY] [_tokens.css](file:///c:/Users/LENOVO/Documents/PPT/Tier_Log/tierlog_v2/resources/css/_tokens.css)

### 2. Performance Layer
Integrate `RecycleScroller` for the feedback list.

#### [MODIFY] [Index.vue](file:///c:/Users/LENOVO/Documents/PPT/Tier_Log/tierlog_v2/resources/js/pages/Consultation/Index.vue)
- Replace static list with `RecycleScroller`.
- Implement reactive highlight logic.

### 3. Neural Actions
Refactor the "Ask AI to Fix" logic into a collaborative neural action.

#### [MODIFY] [Index.vue](file:///c:/Users/LENOVO/Documents/PPT/Tier_Log/tierlog_v2/resources/js/pages/Consultation/Index.vue)
- Update `askAI` to handle specific feedback context.
- Add "Patch" visualization in the chat.

### 4. Visual Polish
Enhance `NeuralNetwork.vue` for reactivity.

#### [MODIFY] [NeuralNetwork.vue](file:///c:/Users/LENOVO/Documents/PPT/Tier_Log/tierlog_v2/resources/js/Components/NeuralNetwork.vue)
- Add mouse-reactive clustering logic.

## Verification Plan

### Automated Tests
- N/A (UI-centric phase)

### Manual Verification
- Verify feedback list scroll performance with 50+ items.
- Verify "Ask AI to Fix" correctly focuses the AI on the specific feedback content.
- Verify the canvas animation reacts to interaction.
