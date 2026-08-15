import 'package:flutter/material.dart';

import '../../core/theme/theme_extensions.dart';

/// Standard branded scaffold: app bar + optional bottom navigation.
class AppScaffold extends StatelessWidget {
  const AppScaffold({
    super.key,
    required this.title,
    required this.body,
    this.actions,
    this.bottomNavIndex,
    this.onBottomNavTap,
  });

  final String title;
  final Widget body;
  final List<Widget>? actions;
  final int? bottomNavIndex;
  final ValueChanged<int>? onBottomNavTap;

  static const _navItems = <(IconData, String)>[
    (Icons.dashboard_outlined, 'Dashboard'),
    (Icons.people_outline, 'Students'),
    (Icons.payments_outlined, 'Fees'),
    (Icons.notifications_outlined, 'Notices'),
  ];

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return Scaffold(
      appBar: AppBar(
        title: Text(title, style: context.typography.title?.copyWith(color: colors.text)),
        actions: actions,
      ),
      body: body,
      bottomNavigationBar: bottomNavIndex == null
          ? null
          : NavigationBar(
              selectedIndex: bottomNavIndex,
              onDestinationSelected: onBottomNavTap,
              destinations: [
                for (final (icon, label) in _navItems)
                  NavigationDestination(icon: Icon(icon), label: label),
              ],
            ),
    );
  }
}
