const { resolve } = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

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
                test: /\.css$/,
                exclude: /node_modules/,
                use: [
                    {
                        loader: MiniCssExtractPlugin.loader,
                        options: { hmr: false }
                    },
                    {
                        loader: 'css-loader',
                        options: { importLoaders: 1 },
                    },
                    {

                        loader: 'postcss-loader',
                        options: {
                            ident: 'postcss',
                            plugins: () => [
                                require('postcss-easy-import')(),
                                require('postcss-mixins')(),
                                require('postcss-calc')(),
                                require('postcss-nested')(),
                                require('postcss-color-mod-function')(),
                                require('postcss-preset-env')(),
                                require('postcss-custom-media')(),
                            ]
                        }
                    }
                ]
            },
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
    },
    plugins: [
        new MiniCssExtractPlugin({
            filename: 'main.css',
        })
    ]
}