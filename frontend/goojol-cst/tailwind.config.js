/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.tsx', './components/**/*.tsx', './feature/**/*.tsx'],
  presets: [require('nativewind/preset')],
  theme: {
    extend: {
      colors: {
        goojol: {
          sky: '#0F1729',
          surface: '#1A2238',
          road: '#E8ECF4',
          coral: '#FF6B4A',
          teal: '#3DDBA8',
          pixel: '#FFD166',
          muted: '#8892A8',
          border: '#2A3550',
        },
      },
    },
  },
  plugins: [],
};
