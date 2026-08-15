import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../domain/dashboard_models.dart';

class DashboardRepository {
  DashboardRepository(this._api);

  final ApiClient _api;

  Future<DashboardSummary> fetchSummary() async {
    final res = await _api.get('/dashboard/summary');
    final data = (res as Map<String, dynamic>)['data'] as Map<String, dynamic>;
    return DashboardSummary.fromJson(data);
  }
}

final dashboardRepositoryProvider = Provider<DashboardRepository>((ref) {
  return DashboardRepository(ref.watch(apiClientProvider));
});
