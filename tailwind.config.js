module.exports = {
  content: [
    './templates/**/*.{html,js}'
  ],
  daisyui: {
    themes: [
      {
        vreco: {

          "primary": "#f27329",

          "secondary": "#d1d5db",

          "accent": "#713f12",

          "neutral": "#3D4451",

          "base-100": "#111111",

          "info": "#3ABFF8",

          "success": "#36D399",

          "warning": "#FBBD23",

          "error": "#F87272",
        },
      },
    ],
  },


  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
}
