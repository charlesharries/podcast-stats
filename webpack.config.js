const { resolve } = require('path');

module.exports = {
    mode: process.env.NODE_ENV === 'production' ? 'production' : 'development',
    entry: './assets/js/app.js',
    output: {
        path: resolve(__dirname, 'static'),
        filename: 'app.js',
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: {
                    loader: 'babel-loader',
                    options: {
                        presets: [['@babel/preset-env', { targets: { node: 10 }}]],
                        plugins: ['@babel/plugin-proposal-class-properties'],
                    }
                }
            }
        ]
    }
}