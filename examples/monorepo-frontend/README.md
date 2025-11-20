# Frontend Monorepo Example

## Overview

This configuration enforces frontend monorepo patterns for React/TypeScript projects.

## Monorepo Structure

```
monorepo/
├── apps/
│   ├── admin/              (Admin Dashboard)
│   │   ├── src/
│   │   │   ├── components/
│   │   │   ├── pages/
│   │   │   ├── hooks/
│   │   │   └── services/
│   │   └── package.json
│   │
│   └── customer/           (Customer App)
│       ├── src/
│       │   ├── components/
│       │   ├── pages/
│       │   ├── hooks/
│       │   └── services/
│       └── package.json
│
└── packages/
    ├── design-system/      (Shared UI Components)
    │   ├── src/
    │   │   └── components/
    │   └── package.json
    │
    ├── utils/              (Shared Utilities)
    └── hooks/              (Shared React Hooks)
```

## Key Principles

- **App Independence**: Apps cannot import from each other
- **Shared Design System**: Common UI components in packages/
- **Shared Utilities**: Reusable hooks and utilities
- **Consistent Naming**: PascalCase components, camelCase hooks

## Naming Conventions

- **Components**: `PascalCase.tsx` (e.g., `Button.tsx`)
- **Pages**: `PascalCase.tsx` (e.g., `HomePage.tsx`)
- **Hooks**: `useCamelCase.ts` (e.g., `useAuth.ts`)
- **Services**: `camelCase.ts` (e.g., `apiService.ts`)

## Usage

```bash
cp .structurelint.yml your-monorepo/
cd your-monorepo
structurelint
```

## Benefits

- Prevents circular dependencies
- Enforces code sharing through design system
- Maintains app independence for deployments
- Consistent structure across apps
