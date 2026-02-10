# MarketPlace

A modern, cross-platform mobile marketplace application built with React Native and Expo.

## Features

- **Product Catalog** - Browse products in a responsive two-column grid layout
- **Product Details** - View detailed information for each product
- **Shopping Cart** - Add, remove, and adjust product quantities with real-time totals
- **Categories** - Browse products by category
- **Order Tracking** - View order history with status indicators (Pending, Shipped, In Transit, Completed)
- **User Authentication** - Secure sign-in/sign-up with email and password
- **Secure Storage** - AES-256 encrypted session storage for enhanced security

## Tech Stack

- **Framework:** [Expo](https://expo.dev) v54 with React Native 0.81
- **Navigation:** [Expo Router](https://docs.expo.dev/router/introduction/) v6 (file-based routing)
- **Backend:** [Supabase](https://supabase.com) (authentication & database)
- **State Management:** [Zustand](https://zustand-demo.pmnd.rs/)
- **Forms:** [React Hook Form](https://react-hook-form.com) + [Zod](https://zod.dev) validation
- **UI Components:** [@rneui/themed](https://reactnativeelements.com/)
- **Language:** TypeScript

## Prerequisites

- [Node.js](https://nodejs.org) (v18 or later recommended)
- [npm](https://www.npmjs.com/) or [yarn](https://yarnpkg.com/)
- [Expo CLI](https://docs.expo.dev/get-started/installation/)
- [Expo Go](https://expo.dev/client) app (for testing on physical devices)

## Getting Started

### 1. Clone the repository

```bash
git clone <repository-url>
cd MarketPlace
```

### 2. Install dependencies

```bash
npm install
```

### 3. Start the development server

```bash
npm start
```

### 4. Run on a device or emulator

- **iOS Simulator:** Press `i` in the terminal or run `npm run ios`
- **Android Emulator:** Press `a` in the terminal or run `npm run android`
- **Web Browser:** Press `w` in the terminal or run `npm run web`
- **Expo Go:** Scan the QR code with the Expo Go app

## Project Structure

```
MarketPlace/
├── assets/                 # Images, icons, and static data
│   ├── images/            # Product and category images
│   ├── types/             # TypeScript type definitions
│   ├── products.ts        # Product data
│   ├── orders.ts          # Order data
│   └── categories.tsx     # Category data
├── src/
│   ├── app/               # Expo Router screens (file-based routing)
│   │   ├── (shop)/        # Main shop screens (home, orders)
│   │   ├── product/       # Product detail screens
│   │   ├── categories/    # Category screens
│   │   ├── auth.tsx       # Authentication screen
│   │   ├── cart.tsx       # Shopping cart screen
│   │   └── _layout.tsx    # Root layout
│   ├── components/        # Reusable UI components
│   ├── lib/               # Utilities and configurations
│   │   └── supabase.ts    # Supabase client setup
│   ├── provider/          # React context providers
│   │   └── auth-provider.tsx
│   └── store/             # Zustand state stores
│       └── cart-store.ts
├── app.json               # Expo configuration
├── package.json           # Dependencies and scripts
└── tsconfig.json          # TypeScript configuration
```

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm start` | Start the Expo development server |
| `npm run android` | Start on Android emulator |
| `npm run ios` | Start on iOS simulator |
| `npm run web` | Start in web browser |

## Environment Setup

The app uses Supabase for authentication and database. The Supabase client is configured in `src/lib/supabase.ts`.

For production deployments, you should:
1. Create your own Supabase project at [supabase.com](https://supabase.com)
2. Update the `supabaseUrl` and `supabaseAnonKey` in `src/lib/supabase.ts`
3. Set up the required database tables (users, products, orders)

## Security Features

- Session tokens are encrypted using AES-256 before storage
- Encryption keys are stored in Expo SecureStore
- Encrypted data is stored in AsyncStorage
- Auto-refresh tokens for seamless authentication

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is private and not licensed for public distribution.
