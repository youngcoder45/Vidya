/// Typed exception mapped from the backend error envelope
/// `{"error": {"code", "message", "details"}}`.
class ApiException implements Exception {
  const ApiException({required this.code, required this.message, this.statusCode, this.details});

  final String code;
  final String message;
  final int? statusCode;
  final Object? details;

  bool get isUnauthorized => statusCode == 401;
  bool get isForbidden => statusCode == 403;
  bool get isRateLimited => statusCode == 429 || code == 'RATE_LIMITED';

  @override
  String toString() => 'ApiException($code): $message';
}

/// Local failures (no network, timeouts, parse errors).
class NetworkException implements Exception {
  const NetworkException(this.message);
  final String message;

  @override
  String toString() => 'NetworkException: $message';
}
