module.exports = {
    default: {
        requireModule: ['godog'],
        require: ['test/functional/steps/*.go'],
        format: ['progress-bar', 'html:cucumber-report.html'],
        formatOptions: { snippetInterface: 'async-await' },
        publishQuiet: true
    }
}; 