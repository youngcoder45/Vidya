import 'package:flutter/material.dart';

import 'semantics/app_semantics.dart';
import 'tokens/app_radii.dart';
import 'tokens/app_spacing.dart';
import 'tokens/app_typography.dart';

/// Derives a full Material [ThemeData] from a semantic role set. Because every
/// widget reads through [AppSemantics]/tokens, changing a token re-themes the
/// entire application — no component hand-tunes colors here.
ThemeData buildAppTheme(AppSemantics s, Brightness brightness) {
  final ColorScheme scheme = ColorScheme(
    brightness: brightness,
    primary: s.primary,
    onPrimary: s.onPrimary,
    secondary: s.info,
    onSecondary: s.surface,
    error: s.danger,
    onError: brightness == Brightness.dark ? s.background : s.surface,
    surface: s.surface,
    onSurface: s.text,
    onSurfaceVariant: s.textMuted,
    outline: s.border,
    surfaceContainerHighest: s.surfaceVariant,
    scrim: s.scrim,
  );

  return ThemeData(
    useMaterial3: true,
    brightness: brightness,
    colorScheme: scheme,
    scaffoldBackgroundColor: s.background,
    textTheme: TextTheme(
      displaySmall: AppTypography.displaySmall.copyWith(color: s.text),
      headlineMedium: AppTypography.headline.copyWith(color: s.text),
      titleLarge: AppTypography.title.copyWith(color: s.text),
      bodyMedium: AppTypography.body.copyWith(color: s.text),
      bodyLarge: AppTypography.bodyStrong.copyWith(color: s.text),
      labelMedium: AppTypography.label.copyWith(color: s.textMuted),
      bodySmall: AppTypography.caption.copyWith(color: s.textMuted),
    ),
    appBarTheme: AppBarTheme(
      backgroundColor: s.surface,
      foregroundColor: s.text,
      elevation: 0,
      centerTitle: false,
    ),
    cardTheme: CardThemeData(
      color: s.surface,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppRadii.md),
        side: BorderSide(color: s.border.withValues(alpha: 0.2)),
      ),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: s.surfaceVariant,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppRadii.sm),
        borderSide: BorderSide.none,
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(AppRadii.sm),
        borderSide: BorderSide(color: s.primary, width: 1.5),
      ),
      contentPadding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.sm,
      ),
    ),
    dividerTheme: DividerThemeData(color: s.border.withValues(alpha: 0.25)),
    snackBarTheme: SnackBarThemeData(
      behavior: SnackBarBehavior.floating,
      backgroundColor: s.text,
      contentTextStyle: AppTypography.body.copyWith(color: s.background),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: s.primary,
        foregroundColor: s.onPrimary,
        minimumSize: const Size.fromHeight(48),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadii.sm)),
        textStyle: AppTypography.bodyStrong,
      ),
    ),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: s.primary,
        side: BorderSide(color: s.primary),
        minimumSize: const Size.fromHeight(48),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadii.sm)),
      ),
    ),
    progressIndicatorTheme: ProgressIndicatorThemeData(color: s.primary),
    extensions: <ThemeExtension<dynamic>>[s],
  );
}

/// Light theme — derived from the light semantic roles.
final ThemeData appLightTheme = buildAppTheme(AppSemantics.light, Brightness.light);

/// Dark theme — derived from the dark semantic roles.
final ThemeData appDarkTheme = buildAppTheme(AppSemantics.dark, Brightness.dark);

/// Convenience: theme for a tenant-branded primary color (light mode).
ThemeData brandedLightTheme(Color brandPrimary) =>
    buildAppTheme(AppSemantics.light.withBrand(brandPrimary), Brightness.light);
