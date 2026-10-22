/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        ch: {
          dark: '#0b0f19',
          card: '#131b2e',
          border: '#1e293b',
          muted: '#334155',
          yellow: '#f59e0b',
          red: '#ef4444',
          accent: '#38bdf8'
        }
      }
    },
  },
  plugins: [],
}
