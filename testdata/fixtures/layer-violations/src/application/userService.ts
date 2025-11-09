import { User } from '../domain/user';

export class UserService {
  getUser() {
    return new User();
  }
}
