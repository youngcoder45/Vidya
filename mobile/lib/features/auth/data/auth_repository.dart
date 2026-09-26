import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/storage/secure_store.dart';
import '../domain/auth_state.dart';

/// Parsed login payload.
class AuthSession {
  const AuthSession({required this.user, required this.schoolId, required this.roles});

  final AuthUser user;
  final String? schoolId;
  final List<String> roles;

  factory AuthSession.fromJson(Map<String, dynamic> json) {
    final user = AuthUser.fromJson(json['user'] as Map<String, dynamic>);
    final roles = (json['roles'] as List<dynamic>? ?? const []).cast<String>();
    return AuthSession(
      user: user,
      schoolId: json['school_id'] as String?,
      roles: roles,
    );
  }
}

class AuthRepository {
  AuthRepository(this._api, this._store);

  final ApiClient _api;
  final SecureStore _store;

  Future<AuthSession> login({required String identifier, required String password}) async {
    final deviceId = await _store.deviceId();
    final res = await _api.post('/auth/login', body: {
      'identifier': identifier,
      'password': password,
      'device': {'device_id': deviceId, 'name': 'Mobile', 'platform': 'android'},
    });
    final data = (res as Map<String, dynamic>)['data'] as Map<String, dynamic>;
    await _store.saveTokens(
      access: data['access_token'] as String,
      refresh: data['refresh_token'] as String,
    );
    return AuthSession.fromJson(data);
  }

  Future<void> logout() async {
    try {
      await _api.post('/auth/logout', body: {'device_id': await _store.deviceId()});
    } catch (_) {
      // best-effort: clear locally regardless
    }
    await _store.clearTokens();
  }

  Future<AuthSession> restoreSession() async {
    // Token exists → assume authenticated until /auth/me proves otherwise.
    try {
      final res = await _api.get('/auth/me');
      final data = (res as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      return AuthSession.fromJson({
        'user': data,
        'school_id': null,
        'roles': const <String>[],
      });
    } on ApiException catch (e) {
      // A rejected token is dead weight; drop it so login starts clean.
      if (e.isUnauthorized) {
        await _store.clearTokens();
      }
      rethrow;
    }
  }
}

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(ref.watch(apiClientProvider), ref.watch(secureStoreProvider));
});
