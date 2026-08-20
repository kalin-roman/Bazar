 # CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
npm install        # install dependencies
npm start           # start Expo dev server (Metro)
npm run android      # start on Android emulator
npm run ios          # start on iOS simulator
npm run web           # start in web browser
```

There is no lint, typecheck, or test script configured in `package.json`, and no ESLint/Jest config files exist in the repo. Use `npx tsc --noEmit` if you need to typecheck manually (TypeScript is strict, extending `expo/tsconfig.base`).

`npm install` needs the root `.npmrc` (`legacy-peer-deps=true`) to succeed unattended — without it, npm's strict resolver ERESOLVEs on `@rneui/themed@4.0.0-rc.8` (a never-stabilized prerelease) peer-requiring `react-native-safe-area-context@"^3 || ^4"` while the project pins `~5.6.0` for `expo-router`/`@react-navigation` v7. Don't try to "fix" this by bumping `@rneui/themed` to the real `5.0.0` release — that trades it for a worse conflict (its peer `@testing-library/jest-native` drags in a `react-test-renderer` version that fights the pinned `react@19.1.0`). The `.npmrc` pin is the intended fix; leave it in place.

`npm audit` currently reports ~19 vulnerabilities (moderate/high), all transitive inside Expo/Metro's own build tooling (`metro`, `@expo/config-plugins`, `postcss`, `uuid`, `xcode`), not in code this app ships. `npm audit fix` (no flags) resolves none of them — the only path is `npm audit fix --force`, which bumps `expo` from `~54.0.31` to `57.0.12` (a 3-major-version jump). Treat that as a deliberate SDK-upgrade task, not something to run reflexively for an audit warning.

`npm run web` requires `react-dom` and `react-native-web` — they're now in `package.json`'s dependencies (added via `npx expo install react-dom react-native-web`), so this should just work.

## Architecture

This is an Expo (v54) / React Native (0.81) marketplace app using **Expo Router v6** file-based routing. The Expo config (`app.json`) names the app "MarketPlace" even though the repo/folder is "Bazar" — these are the same project.

### Entry point

`package.json` sets `"main": "expo-router/entry"`, so Expo Router's own entry file drives the app and reads routes from `src/app/`. The `index.ts` at the repo root (which manually calls `registerRootComponent` on `src/app/(shop)`) is not the active entry point — don't assume changes there affect app boot.

### Routing structure (`src/app/`)

- `_layout.tsx` — root layout: `ToastProvider` > `AuthProvider` > `Stack` with screens for `(shop)`, `product`, `categories`, `cart` (modal), `auth`.
- `(shop)/_layout.tsx` — tab navigator (Shop, Orders). The auth gate is currently **disabled** (see Auth section below) — it renders the tabs unconditionally with no `useAuth()`/`Redirect` check, so every screen is reachable without signing in.
- `(shop)/index.tsx` — product grid (home).
- `(shop)/orders/` — order list/detail, backed by static mock data.
- `product/[slug].tsx`, `categories/[slug].tsx` — detail screens driven by route params, looked up against static mock arrays (not fetched by id).
- `cart.tsx` — presented as a modal from the root stack.

### Auth (`src/provider/auth-provider.tsx`, `src/lib/supabase.ts`)

- **Currently disconnected.** `AuthProvider` (`src/provider/auth-provider.tsx`) is a stub: it makes no Supabase calls at all and always provides `{ session: null, mounting: false, user: null }` via `useAuth()`. This was done deliberately so every screen is browsable locally without a login. `src/lib/supabase.ts` and `src/app/auth.tsx` (the login/sign-up screen) are untouched and still fully functional — nothing currently imports/calls them.
- To reconnect real auth: restore `AuthProvider` to load the Supabase session on mount, subscribe to `supabase.auth.onAuthStateChange`, and fetch the matching row from the Supabase `users` table (that's the original, git-historical implementation) — and put back the `useAuth()` + `Redirect to /auth` gate in `(shop)/_layout.tsx` that was removed.
- The Supabase client uses a custom `LargeSecureStore` adapter: it AES-256-encrypts (via `aes-js`) the session blob before writing it to `AsyncStorage`, while the AES key itself lives in `expo-secure-store` (worked around because SecureStore has a 2048-byte value limit).
- The Supabase URL and anon key are hardcoded directly in `src/lib/supabase.ts` rather than sourced from env/`expo-constants` — there is no `.env` file in this project.

### Data layer

Product, category, and order data are **static mock arrays** in `assets/` (`products.ts`, `categories.tsx`, `orders.ts`), typed via `assets/types/`. Screens import these directly (e.g. `PRODUCTS.find(p => p.slug === slug)`) rather than querying Supabase. Supabase was previously wired up only for auth and the `users` table lookup, and that path is currently disconnected too (see Auth section) — treat any "fetch products/orders from Supabase" work as a net-new integration, not a bug fix, and check whether corresponding tables/schemas exist first.

Note `assets/` lives at the repo root, outside `src/`, so imports from deep route files use long relative paths like `../../../assets/products`.

`assets/orders.ts`'s mock `Order.items` entries follow the same shape as `assets/products.ts` (`imagesUrl` as `require(...)` calls, plus `category` and `maxQuantity`) — they were previously missing those fields/using raw strings, which `tsc --noEmit` only partially caught (it flagged the `imagesUrl` type mismatch but not the missing `category`/`maxQuantity` properties, for reasons not fully understood — don't assume a clean `tsc` run means mock data fully matches its declared type).

### State

`src/store/cart-store.ts` is a Zustand store holding cart items entirely client-side (no persistence, no Supabase sync). It clamps quantities against each product's `maxQuantity` from the mock product data.
