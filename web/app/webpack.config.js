const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

module.exports = {
  entry: './src/app.ts',
  devtool: 'inline-source-map',
  plugins: [
    new MiniCssExtractPlugin({
      filename: 'app.css',
      chunkFilename: '[id].css',
      ignoreOrder: false
    })
  ],
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: 'ts-loader',
        exclude: /node_modules/
      },
      {
        test: /\.(png|svg|jpg|jpeg|gif)$/i,
        type: 'asset/inline'
      },
      {
        test: /\.s[ac]ss$/i,
        use: [
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              publicPath: path.resolve(__dirname, '../static/assets')
            }
          },
          'css-loader',
          'postcss-loader',
          'sass-loader'
        ]
      }
    ]
  },
  resolve: {
    extensions: ['.ts', '.js', '.scss']
  },
  output: {
    filename: 'app.js',
    path: path.resolve(__dirname, '../static/assets')
  }
};
