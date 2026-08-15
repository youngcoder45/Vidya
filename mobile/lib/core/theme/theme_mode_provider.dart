import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Theme mode with persistence (system default, user-overridable).
class ThemeModeController extends Notifier<ThemeMode> {
  static const _prefKey = 'theme.mode';

  @override
  ThemeMode build() {
    final prefs = SharedPreferences.getInstance();
    prefs.then((p) {
      final raw = p.getString(_prefKey);
      if (raw != null) {
        state = ThemeMode.values.firstWhere(
          (m) => m.name == raw,
          orElse: () => ThemeMode.system,
        );
      }
    });
    return ThemeMode.system;
  }

  Future<void> setMode(ThemeMode mode) async {
    state = mode;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_prefKey, mode.name);
  }
}

final themeModeProvider = NotifierProvider<ThemeModeController, ThemeMode>(ThemeModeController.new);
