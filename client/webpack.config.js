const path = require('path');

module.exports = {
    entry: './src/app.ts',
    devtool: 'inline-source-map',
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
        ],
    },
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
        alias: {
            '@domain': path.resolve(__dirname, 'src/domain'),
            '@services': path.resolve(__dirname, 'src/services'),
            '@stores': path.resolve(__dirname, 'src/stores'),
            '@types': path.resolve(__dirname, 'src/domain/types'),
            '@utilities': path.resolve(__dirname, 'src/utilities')
        }
    },
    output: {
        filename: 'app.js',
        path: path.resolve(__dirname, '../public/assets'),
    },
};
