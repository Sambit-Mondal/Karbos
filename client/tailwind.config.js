/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",

    // Or if using `src` directory:
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        karbos: {
          navy: "#1A1B41",
          indigo: "#292E6F",
          "blue-purple": "#525FB0",
          lavender: "#A6B1E1",
          "light-blue": "#D6DFFF",
        },
      },
    },
  },
  plugins: [],
};
