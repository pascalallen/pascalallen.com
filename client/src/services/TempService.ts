import request from '@utilities/request';
import HttpMethod from '@domain/constants/HttpMethod';
import AuthStore from '@stores/AuthStore';

class TempService {
  private readonly authStore: AuthStore;

  constructor(authStore: AuthStore) {
    this.authStore = authStore;
  }

  public async temp() {
    const response = await request.send({
      method: HttpMethod.GET,
      uri: '/api/v1/temp',
      options: { auth: true, authStore: this.authStore }
    });

    return response.body.data;
  }
}

export default TempService;
