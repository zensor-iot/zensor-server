const path = require('node:path')

const designSystem = path.resolve(__dirname, '../docs/design-system')

module.exports = {
  presets: [require(path.join(designSystem, 'tailwind.preset.js'))],
  content: ['./index.html', './src/**/*.{js,jsx}'],
}
