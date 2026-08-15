import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/dashboard_repository.dart';
import '../domain/dashboard_models.dart';

/// Server state for the admin dashboard; invalidated after mutations
/// (e.g., fee payment) to refetch.
final dashboardSummaryProvider = FutureProvider<DashboardSummary>((ref) async {
  return ref.watch(dashboardRepositoryProvider).fetchSummary();
});
