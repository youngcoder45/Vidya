import 'package:go_router/go_router.dart';

import '../../features/auth/domain/auth_state.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/dashboard/presentation/dashboard_screen.dart';

/// Builds the router for the current auth state. Rebuilt when auth changes
/// (see app.dart) — a single redirect rule keeps route guards in one place.
///
/// Deep links (notification taps) will carry route + params here later.
GoRouter buildRouter(AuthState auth) {
  return GoRouter(
    initialLocation: '/dashboard',
    routes: [
      GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
      GoRoute(path: '/dashboard', builder: (_, __) => const DashboardScreen()),
    ],
    redirect: (context, state) {
      final isLogin = state.matchedLocation == '/login';
      if (!auth.isAuthenticated) {
        return isLogin ? null : '/login';
      }
      if (isLogin) return '/dashboard';
      return null;
    },
  );
}
