/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
  ],
  corePlugins: {
    preflight: false,
  },
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#3d9970',
          dark: '#2f7a58',
          light: '#e8f7f0',
        },
      },
    },
  },
  plugins: [],
};
