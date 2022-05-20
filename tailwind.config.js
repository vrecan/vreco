module.exports = {
  content: [
    './templates/**/*.{html,js}'
  ],
  daisyui: {
    themes: ["halloween", "luxary"],
  },  

  
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
}
