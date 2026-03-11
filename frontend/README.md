# Frontend — Sprout Accounting

Next.js 15 + React 19 + Tailwind CSS v4 + shadcn/ui frontend for the Sprout Accounting system.

## Getting Started

### Install dependencies

```bash
npm install
```

### Initialize additional shadcn/ui components (optional)

```bash
npx shadcn@latest add <component-name>
```

For example:

```bash
npx shadcn@latest add table dialog select badge
```

### Run the development server

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Tech Stack

- **Framework**: [Next.js 15](https://nextjs.org/) (App Router, Turbopack)
- **Language**: TypeScript 5
- **UI Library**: React 19
- **Styling**: Tailwind CSS v4
- **Component Library**: [shadcn/ui](https://ui.shadcn.com/) (New York style)
- **Icons**: [Lucide React](https://lucide.dev/)

## Project Structure

```
src/
├── app/              # Next.js App Router pages & layouts
│   ├── globals.css   # Tailwind CSS + CSS variables (theme)
│   ├── layout.tsx    # Root layout
│   ├── page.tsx      # Home page
│   ├── loading.tsx   # Loading fallback
│   └── not-found.tsx # 404 page
├── components/
│   └── ui/           # shadcn/ui components
│       ├── button.tsx
│       ├── card.tsx
│       └── input.tsx
├── hooks/            # Custom React hooks
│   └── use-mobile.ts
└── lib/
    └── utils.ts      # Utility functions (cn helper)
```

## Scripts

| Command          | Description                        |
| ---------------- | ---------------------------------- |
| `npm run dev`    | Start dev server with Turbopack    |
| `npm run build`  | Build for production               |
| `npm run start`  | Start production server            |
| `npm run lint`   | Run ESLint                         |
