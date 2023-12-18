/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./view/**/*.{templ,go}"],
  theme: {
    screens: {
      mobile: "360px",
      tablet: "640px",
      laptop: "1024px",
      desktop: "1280px",
    },
    extend: {},
  },
  corePlugins: {
    preflight: false,
  },
  plugins: [],
};
