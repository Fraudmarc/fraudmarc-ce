const apiEndpoint = '';
const userPoolId = '';
const userPoolClientId = '';

const apiPath: string = '/api';

export const commonEnvironment = {
  apiBaseUrl: `${apiEndpoint}${apiPath}`,
  Cognito: {
    userPoolId,
    userPoolClientId,
  },
};
