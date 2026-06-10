# TierLog Web

Frontend web berbasis **Laravel + Inertia.js + Vue 3 + TypeScript** (VILT Stack).

## Tech Stack

- **Laravel 13** — Backend framework, routing, Inertia server-side rendering
- **Inertia.js** — SPA-like experience without API layer (bridges Laravel ↔ Vue)
- **Vue 3** — Frontend framework with Composition API (`<script setup>`)
- **TypeScript** — Type-safe development
- **Pinia** — State management (replaces React Context/Zustand)
- **Tailwind CSS v4** — Utility-first styling
- **Vite 8** — Build tool and dev server

## Menjalankan

1. Copy `.env.example` menjadi `.env`
2. Generate app key:
   ```bash
   php artisan key:generate
   ```
3. Install dependency PHP:
   ```bash
   composer install
   ```
4. Install dependency Node:
   ```bash
   npm install
   ```
5. Jalankan development server:
   ```bash
   npm run dev
   ```
6. Jalankan Laravel server (di terminal lain):
   ```bash
   php artisan serve
   ```

Frontend akan memanggil backend Go pada `GO_API_URL` (default: `http://127.0.0.1:8080`).

## Struktur

```
resources/js/
├── app.ts                  # Entry point (Inertia + Vue + Pinia)
├── types.ts                # TypeScript type definitions
├── stores/auth.ts          # Pinia auth store (login, register, API wrapper)
├── composables/useApi.ts   # API composable
├── components/             # Reusable Vue components
│   ├── icons.ts            # SVG icon components
│   ├── NavBar.vue          # Navigation bar
│   ├── RequireAuth.vue     # Auth guard
│   ├── UiCard.vue          # Frosted glass card
│   ├── UiButton.vue        # Action button
│   ├── UiBadge.vue         # Status badge
│   ├── UiField.vue         # Form input
│   ├── UiHeading.vue       # Section heading
│   ├── UiPage.vue          # Page wrapper
│   └── UiStatCard.vue      # Metric card
└── pages/                  # Inertia page components
    ├── Welcome.vue         # Landing page
    ├── Login.vue           # Login page
    ├── Register.vue        # Registration page
    ├── Dashboard.vue       # Student/Lecturer dashboard
    ├── Consultations.vue   # Consultation workspace
    ├── Archive.vue         # Session archive
    ├── LecturerDashboard.vue # Lecturer supervisor portal
    └── settings/
        ├── Profile.vue     # Profile settings
        ├── Security.vue    # Password settings
        └── AiGateway.vue   # AI model configuration
```
