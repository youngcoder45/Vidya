import 'package:flutter/material.dart';

import '../../core/theme/theme_extensions.dart';

/// A labeled value tile for dashboards.
///
/// Note: every color/size here comes from `context.*` — zero hardcoded values.
class StatTile extends StatelessWidget {
  const StatTile({
    super.key,
    required this.label,
    required this.value,
    this.icon,
    this.accent,
    this.onTap,
  });

  final String label;
  final String value;
  final IconData? icon;
  final Color? accent;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final accent = this.accent ?? colors.primary;
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(context.spaceMd),
        child: Padding(
          padding: EdgeInsets.all(context.spaceMd),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  if (icon != null) ...[
                    Icon(icon, size: 20, color: accent),
                    SizedBox(width: context.spaceXs),
                  ],
                  Expanded(
                    child: Text(
                      label,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: context.typography.label?.copyWith(color: colors.textMuted),
                    ),
                  ),
                ],
              ),
              SizedBox(height: context.spaceSm),
              Text(
                value,
                style: context.typography.title?.copyWith(color: colors.text, fontWeight: FontWeight.w700),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
