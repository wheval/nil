module.exports = function (context, options) {
  return {
    name: 'custom-loaders',
    configureWebpack(config, isServer) {
      return {
        module: {
          rules: [
            {
              test: /\.svg$/i,
              use: ['@svgr/webpack'],
            },
          ],
        },
      };
    },
  };
};