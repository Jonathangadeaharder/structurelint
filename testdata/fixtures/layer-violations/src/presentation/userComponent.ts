import { UserService } from '../application/userService';
import { User } from '../domain/user';

export function UserComponent() {
  const service = new UserService();
  return service.getUser();
}
