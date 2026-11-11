const fs = require('fs');
const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');

class RemoveStaleCompressedPlugin {
  apply(compiler) {
    compiler.hooks.afterEmit.tap('RemoveStaleCompressedPlugin', compilation => {
      for (const name of Object.keys(compilation.assets)) {
        for (const suffix of ['.br', '.gz']) {
          fs.rmSync(path.join(compiler.options.output.path, name + suffix), { force: true });
        }
      }
    });
  }
}

module.exports = (_env, argv) => ({
  entry: {
    'app': './src/app.ts',
    ...(argv.mode === 'production' ? {} : { 'dev-shell': './src/dev-shell.ts' }),
    'theme': './src/theme-boot.ts'
  },
  devtool: argv.mode === 'production' ? 'source-map' : 'inline-source-map',
  plugins: [
    new MiniCssExtractPlugin({ filename: '[name].css' }),
    ...(argv.mode === 'production' ? [] : [new RemoveStaleCompressedPlugin()])
  ],
  module: {
    rules: [
      {
        test: /\.ts$/,
        use: { loader: 'ts-loader', options: { configFile: 'tsconfig.json' } },
        exclude: /node_modules/
      },
      {
        test: /\.css$/i,
        use: [MiniCssExtractPlugin.loader, 'css-loader']
      }
    ]
  },
  optimization: {
    minimizer: ['...', new CssMinimizerPlugin()]
  },
  resolve: {
    extensions: ['.ts', '.js']
  },
  output: {
    filename: '[name].js',
    path: path.resolve(__dirname, '../static/assets')
  }
});
