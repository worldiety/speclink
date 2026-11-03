import type UserPermissions from '@/domain/types/UserPermissions.ts';

export default class AuthService {
  private readonly permissions: UserPermissions;

  constructor(params: { permissions: UserPermissions }) {
    this.permissions = params.permissions;
  }

  public audit(permission: keyof UserPermissions): void {
    if (!this.permissions[permission]) {
      throw new Error('Permission denied: ' + permission);
    }
  }
}
