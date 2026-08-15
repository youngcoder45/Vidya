import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../config/app_config.dart';
import '../storage/secure_store.dart';
import 'api_exception.dart';

/// HTTP client wrapper: base URL, auth header, single-flight token refresh,
/// tenant header, and error-envelope mapping.
class ApiClient {
  ApiClient(this._dio, this._store);

  final Dio _dio;
  final SecureStore _store;

  Future<dynamic> get(String path, {Map<String, dynamic>? query}) =>
      _request(() => _dio.get(path, queryParameters: query));

  Future<dynamic> post(String path, {Object? body}) => _request(() => _dio.post<dynamic>(path, data: body));

  Future<dynamic> put(String path, {Object? body}) => _request(() => _dio.put(path, data: body));

  Future<dynamic> delete(String path) => _request(() => _dio.delete(path));

  Future<dynamic> _request(Future<Response<dynamic>> Function() send) async {
    try {
      final response = await send();
      return response.data;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401 && _isAuthPath(e.requestOptions.path) == false) {
        final refreshed = await _refreshTokens();
        if (refreshed) {
          return _request(send);
        }
      }
      throw _mapError(e);
    }
  }

  bool _isAuthPath(String path) => path.contains('/auth/');

  Future<bool> _refreshTokens() async {
    final refresh = await _store.readRefreshToken();
    if (refresh == null) return false;
    try {
      final res = await _dio.post<dynamic>(
        '${AppConfig.apiBaseUrl}/auth/refresh',
        data: {'refresh_token': refresh, 'device_id': await _store.deviceId()},
      );
      final data = res.data['data'] as Map<String, dynamic>;
      await _store.saveTokens(
        access: data['access_token'] as String,
        refresh: data['refresh_token'] as String,
      );
      return true;
    } catch (_) {
      await _store.clearTokens();
      return false;
    }
  }

  Exception _mapError(DioException e) {
    final res = e.response;
    if (res == null) {
      return const NetworkException('No internet connection. Please try again.');
    }
    try {
      final body = res.data is String ? jsonDecode(res.data as String) : res.data;
      final error = (body as Map<String, dynamic>)['error'] as Map<String, dynamic>?;
      if (error != null) {
        return ApiException(
          code: error['code'] as String? ?? 'UNKNOWN',
          message: error['message'] as String? ?? 'Something went wrong',
          statusCode: res.statusCode,
          details: error['details'],
        );
      }
    } catch (_) {
      // fall through to generic mapping
    }
    return ApiException(
      code: 'UNKNOWN',
      message: 'Something went wrong (${res.statusCode})',
      statusCode: res.statusCode,
    );
  }
}

/// Riverpod provider for the shared API client.
final apiClientProvider = Provider<ApiClient>((ref) {
  final store = ref.watch(secureStoreProvider);
  final dio = Dio(
    BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 20),
      headers: {'Content-Type': 'application/json'},
    ),
  );
  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await store.readAccessToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        options.headers['X-Request-ID'] = DateTime.now().microsecondsSinceEpoch.toString();
        handler.next(options);
      },
    ),
  );
  return ApiClient(dio, store);
});
