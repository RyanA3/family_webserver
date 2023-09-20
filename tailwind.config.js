/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: "jit",
  purge: ["./views/**/*.{html,js,pug}"],
  content: ["./views/**/*.{html,js,pug}"],
  theme: {
    extend: {},
  },
  plugins: [],
}

