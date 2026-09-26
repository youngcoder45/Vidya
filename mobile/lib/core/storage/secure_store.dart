import 'dart:math';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Persists tokens (secure storage) and lightweight prefs (shared prefs).
class SecureStore {
  SecureStore(this._secure, this._prefs);

  static const _accessKey = 'auth.access_token';
  static const _refreshKey = 'auth.refresh_token';
  static const _deviceIdKey = 'device.id';

  final FlutterSecureStorage _secure;
  final SharedPreferences _prefs;

  Future<void> saveTokens({required String access, required String refresh}) async {
    await _secure.write(key: _accessKey, value: access);
    await _secure.write(key: _refreshKey, value: refresh);
  }

  Future<String?> readAccessToken() => _secure.read(key: _accessKey);
  Future<String?> readRefreshToken() => _secure.read(key: _refreshKey);

  Future<void> clearTokens() async {
    await _secure.delete(key: _accessKey);
    await _secure.delete(key: _refreshKey);
  }

  Future<String> deviceId() async {
    final existing = _prefs.getString(_deviceIdKey);
    if (existing != null) return existing;
    // Random, not derived from the clock: a predictable device id is spoofable.
    final rnd = Random.secure();
    final id = List<int>.generate(16, (_) => rnd.nextInt(256))
        .map((b) => b.toRadixString(16).padLeft(2, '0'))
        .join();
    await _prefs.setString(_deviceIdKey, id);
    return id;
  }
}

final secureStoreProvider = Provider<SecureStore>((ref) {
  throw UnimplementedError('initialized in main() before runApp');
});
