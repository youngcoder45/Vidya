/// Compile-time / runtime configuration.
///
/// Override with `--dart-define=API_BASE_URL=…` per environment.
class AppConfig {
  AppConfig._();

  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080/api/v1', // Android emulator → host
  );

  static const String razorpayKeyId = String.fromEnvironment('RAZORPAY_KEY_ID');

  static const String appName = 'Vidya';
}
