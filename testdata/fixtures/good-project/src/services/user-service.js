import { User } from '../domain/user.js';

export class UserService {
  createUser(name) {
    return new User(name);
  }
}
