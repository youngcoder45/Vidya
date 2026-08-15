import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_exception.dart';
import '../data/auth_repository.dart';
import '../domain/auth_state.dart';

/// Owns the app-wide auth state and drives login/logout/restore.
class AuthController extends Notifier<AuthState> {
  @override
  AuthState build() {
    _restore();
    return const AuthState.unknown();
  }

  AuthRepository get _repo => ref.read(authRepositoryProvider);

  Future<void> _restore() async {
    try {
      final session = await _repo.restoreSession();
      state = AuthState.authenticated(
        user: session.user,
        schoolId: session.schoolId,
        roles: session.roles,
      );
    } catch (_) {
      state = const AuthState.unauthenticated();
    }
  }

  Future<void> login({required String identifier, required String password}) async {
    state = const AuthState.unknown(); // loading
    try {
      final session = await _repo.login(identifier: identifier, password: password);
      state = AuthState.authenticated(
        user: session.user,
        schoolId: session.schoolId,
        roles: session.roles,
      );
    } on ApiException catch (e) {
      state = const AuthState.unauthenticated();
      rethrow;
    } catch (e) {
      state = const AuthState.unauthenticated();
      rethrow;
    }
  }

  Future<void> logout() async {
    await _repo.logout();
    state = const AuthState.unauthenticated();
  }
}

final authControllerProvider =
    NotifierProvider<AuthController, AuthState>(AuthController.new);
