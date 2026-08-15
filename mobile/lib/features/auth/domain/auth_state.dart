/// Authenticated user summary returned by `/auth/login` and `/auth/me`.
class AuthUser {
  const AuthUser({
    required this.id,
    required this.fullName,
    this.email,
    this.phone,
    this.avatarUrl,
  });

  final String id;
  final String fullName;
  final String? email;
  final String? phone;
  final String? avatarUrl;

  factory AuthUser.fromJson(Map<String, dynamic> json) => AuthUser(
        id: json['id'] as String,
        fullName: json['full_name'] as String? ?? 'User',
        email: json['email'] as String?,
        phone: json['phone'] as String?,
        avatarUrl: json['avatar_url'] as String?,
      );
}

/// The auth state of the app.
enum AuthStatus { unknown, unauthenticated, authenticated }

class AuthState {
  const AuthState({required this.status, this.user, this.schoolId, this.roles = const []});

  final AuthStatus status;
  final AuthUser? user;
  final String? schoolId;
  final List<String> roles;

  bool get isAuthenticated => status == AuthStatus.authenticated;

  const AuthState.unknown() : this(status: AuthStatus.unknown);
  const AuthState.unauthenticated() : this(status: AuthStatus.unauthenticated);

  factory AuthState.authenticated({
    required AuthUser user,
    required String? schoolId,
    required List<String> roles,
  }) =>
      AuthState(status: AuthStatus.authenticated, user: user, schoolId: schoolId, roles: roles);
}
