// VIOLATION: domain should not import from application
import { UserService } from '../application/userService';

export class User {
  name: string;
}
