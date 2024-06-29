/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./www/**/*.html", "./www/**/*.templ", "./www/**/*.go"],
  theme: {
    extend: {},
  },
  darkMode: ["class", "[data-theme='dark']"],
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["garden", "dark"],
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
    logs: true,
    themeRoot: ":root",
  },
};
