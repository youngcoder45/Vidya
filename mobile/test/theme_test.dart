import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:vidya_mobile/core/theme/app_theme.dart';
import 'package:vidya_mobile/core/theme/semantics/app_semantics.dart';
import 'package:vidya_mobile/core/theme/tokens/app_colors.dart';
import 'package:vidya_mobile/core/theme/tokens/app_spacing.dart';

void main() {
  test('light semantics derive from tokens', () {
    expect(AppSemantics.light.primary, AppColors.primary);
    expect(AppSemantics.light.surface, AppColors.surface);
    expect(AppSemantics.light.danger, AppColors.danger);
  });

  test('dark mode is a distinct semantic set', () {
    expect(AppSemantics.dark.background, isNot(AppSemantics.light.background));
    expect(AppSemantics.dark.text, isNot(AppSemantics.light.text));
  });

  test('ThemeData is derived from semantics (single source of truth)', () {
    final theme = appLightTheme;
    expect(theme.scaffoldBackgroundColor, AppSemantics.light.background);
    expect(theme.colorScheme.primary, AppSemantics.light.primary);
    expect(theme.extension<AppSemantics>(), AppSemantics.light);
  });

  test('changing a token changes the derived theme', () {
    // brandedLightTheme overrides primary only — everything else stays.
    final brand = brandedLightTheme(const Color(0xFF123456));
    expect(brand.colorScheme.primary, const Color(0xFF123456));
    expect(brand.scaffoldBackgroundColor, AppSemantics.light.background);
  });

  test('spacing scale is consistent', () {
    expect(AppSpacing.xs < AppSpacing.sm, isTrue);
    expect(AppSpacing.sm < AppSpacing.md, isTrue);
    expect(AppSpacing.md < AppSpacing.lg, isTrue);
  });
}
