module.exports = {
    apps: [
        {
            name: "baipiao",
            exec_mode: "fork",
            watch: 'config.toml',
            interpreter: "./baipiao",
            script: '.',
        }
    ]
};
