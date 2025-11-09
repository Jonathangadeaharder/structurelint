import { UserService } from './services/user-service.js';

const service = new UserService();
const user = service.createUser('John');
console.log(user);
