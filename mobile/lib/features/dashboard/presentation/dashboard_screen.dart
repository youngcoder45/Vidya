import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/theme_extensions.dart';
import '../../../shared/widgets/app_scaffold.dart';
import '../../../shared/widgets/stat_tile.dart';
import '../../auth/presentation/auth_controller.dart';
import '../domain/dashboard_models.dart';
import 'dashboard_controller.dart';

final _inr = NumberFormat.currency(locale: 'en_IN', symbol: '₹', decimalDigits: 0);

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summary = ref.watch(dashboardSummaryProvider);
    final user = ref.watch(authControllerProvider).user;

    return AppScaffold(
      title: 'Dashboard',
      bottomNavIndex: 0,
      onBottomNavTap: (_) {},
      actions: [
        IconButton(
          tooltip: 'Logout',
          icon: const Icon(Icons.logout),
          onPressed: () => ref.read(authControllerProvider.notifier).logout(),
        ),
      ],
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(dashboardSummaryProvider.future),
        child: ListView(
          padding: EdgeInsets.all(context.spaceMd),
          children: [
            Text(
              'Welcome, ${user?.fullName ?? 'there'} 👋',
              style: context.typography.headlineMedium?.copyWith(color: context.colors.text),
            ),
            SizedBox(height: context.spaceMd),
            summary.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Text(
                'Could not load dashboard: $e',
                style: context.typography.bodyMedium?.copyWith(color: context.colors.danger),
              ),
              data: (s) => _SummaryGrid(summary: s),
            ),
          ],
        ),
      ),
    );
  }
}

class _SummaryGrid extends StatelessWidget {
  const _SummaryGrid({required this.summary});

  final DashboardSummary summary;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final attendance = summary.attendanceTodayPercent;
    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: StatTile(
                label: 'Students',
                value: '${summary.students}',
                icon: Icons.people_outline,
                accent: colors.info,
              ),
            ),
            SizedBox(width: context.spaceSm),
            Expanded(
              child: StatTile(
                label: 'Teachers',
                value: '${summary.teachers}',
                icon: Icons.school_outlined,
                accent: colors.primary,
              ),
            ),
          ],
        ),
        SizedBox(height: context.spaceSm),
        Row(
          children: [
            Expanded(
              child: StatTile(
                label: 'Attendance today',
                value: attendance == null ? '—' : '${attendance.toStringAsFixed(1)}%',
                icon: Icons.fact_check_outlined,
                accent: colors.success,
              ),
            ),
            SizedBox(width: context.spaceSm),
            Expanded(
              child: StatTile(
                label: 'Pending fees',
                value: '${summary.pendingFeesCount} · ${_inr.format(summary.pendingFeesAmountInr)}',
                icon: Icons.payments_outlined,
                accent: colors.feeAccent,
              ),
            ),
          ],
        ),
        SizedBox(height: context.spaceSm),
        Row(
          children: [
            Expanded(
              child: StatTile(
                label: 'Collected (month)',
                value: _inr.format(summary.revenueCollectedInr),
                icon: Icons.trending_up,
                accent: colors.success,
              ),
            ),
            SizedBox(width: context.spaceSm),
            Expanded(
              child: StatTile(
                label: 'Collection rate',
                value: '${summary.collectionRate.toStringAsFixed(1)}%',
                icon: Icons.pie_chart_outline,
                accent: colors.info,
              ),
            ),
          ],
        ),
      ],
    );
  }
}
