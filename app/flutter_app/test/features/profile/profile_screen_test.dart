import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/gamification/presentation/gamification_demo_controller.dart';
import 'package:flutter_app/features/profile/models/profile_customization.dart';
import 'package:flutter_app/features/profile/presentation/profile_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/theme/meloop_theme_controller.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  tearDown(() {
    MeloopThemeController.instance.resetToDefault();
    GamificationDemoController.instance.level = 3;
    GamificationDemoController.instance.xp = 70;
  });

  testWidgets('shows persisted author posts instead of generated samples', (
    tester,
  ) async {
    final dio = Dio();
    dio.httpClientAdapter = _ProfilePostsAdapter((options) {
      expect(options.method, 'GET');
      expect(options.uri.path, '/api/v1/posts/author/author-1');
      return ResponseBody.fromString(
        jsonEncode([
          {
            'id': 'post-1',
            'author_id': 'author-1',
            'content': 'Publicación guardada en la base de datos',
            'created_at': '2026-11-24T20:00:00Z',
          },
        ]),
        200,
        headers: {
          Headers.contentTypeHeader: ['application/json'],
        },
      );
    });
    final client = ApiClient(dio: dio);
    addTearDown(client.close);

    await tester.pumpWidget(
      MaterialApp(
        home: ProfileScreen(userId: 'author-1', apiClient: client),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      find.text('Publicación guardada en la base de datos'),
      findsOneWidget,
    );
    expect(find.textContaining('Publicación número'), findsNothing);
  });

  testWidgets(
    'desktop profile editor customizes global style, color, layout, and stickers',
    (tester) async {
      tester.view.physicalSize = const Size(1600, 1000);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(
        const MaterialApp(home: ProfileScreen(insideDesktopShell: true)),
      );
      await tester.pumpAndSettle();

      // Default profile matches Mockup 1
      expect(find.text('Sobre mí:'), findsOneWidget);
      expect(find.text('Artistas que me gustan'), findsOneWidget);
      expect(find.text('Canción favorita'), findsOneWidget);
      expect(find.text('Amigos:'), findsOneWidget);
      expect(find.text('Editar perfil'), findsOneWidget);

      // Open live Profile Editor studio
      await tester.tap(find.text('Editar perfil'));
      await tester.pumpAndSettle();
      expect(find.text('Estudio de perfil'), findsOneWidget);

      // Level 3 already unlocks Plano cuadrado, Rosa Pastel, and Cuadrícula pastel
      await tester.tap(find.text('Plano cuadrado'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Rosa Pastel'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Rosa Pastel'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Cuadrícula pastel'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Cuadrícula pastel'));
      await tester.pumpAndSettle();

      expect(
        MeloopThemeController.instance.appStyle,
        MeloopAppStyle.flatSquare,
      );
      expect(
        MeloopThemeController.instance.colorTheme,
        MeloopColorTheme.sakuraPink,
      );
      expect(
        MeloopThemeController.instance.backgroundPattern,
        MeloopBackgroundPattern.pastelGrid,
      );

      // Hide friends card in Diseño tab (Mockup 2 behavior)
      await tester.tap(find.text('Diseño'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Mostrar lista de amigos'));
      await tester.pumpAndSettle();
      expect(find.text('Amigos:'), findsNothing);

      // Add a sticker in Stickers tab
      await tester.tap(find.text('Stickers'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Gatito curioso'));
      await tester.pumpAndSettle();
      expect(MeloopThemeController.instance.config.stickers.length, 1);
    },
  );

  test('ProfileCustomizationConfig serializes and enforces level limits', () {
    const config = ProfileCustomizationConfig(
      appStyle: MeloopAppStyle.frutigerAero,
      colorTheme: MeloopColorTheme.sakuraPink,
      backgroundPattern: MeloopBackgroundPattern.aeroBubbles,
      borderStyle: ProfileBorderStyle.goldenFrame,
      pinId: 'pin_sakura',
      mood: 'Inspirado',
      layout: ProfileLayoutConfig(
        preset: ProfileLayoutPreset.showcase,
        showFriends: false,
      ),
      stickers: [
        ProfileStickerPlacement(
          id: 's1',
          stickerId: 'cat_shy',
          x: 0.25,
          y: 0.35,
          scale: 1.15,
          rotation: 0.12,
        ),
        ProfileStickerPlacement(
          id: 's2',
          stickerId: 'cloud_blush',
          x: 0.72,
          y: 0.41,
        ),
      ],
    );

    final decoded = ProfileCustomizationConfig.fromJson(
      jsonDecode(jsonEncode(config.toJson())) as Map<String, dynamic>,
    );
    expect(decoded.schemaVersion, 1);
    expect(decoded.appStyle, MeloopAppStyle.frutigerAero);
    expect(decoded.colorTheme, MeloopColorTheme.sakuraPink);
    expect(decoded.layout.showFriends, isFalse);
    expect(decoded.stickers.length, 2);

    // Sanitizing for Level 3 strips level 10/6 unlocks & caps stickers to 1
    final sanitized = decoded.sanitizedForLevel(3);
    expect(sanitized.appStyle, MeloopAppStyle.paper);
    expect(sanitized.colorTheme, MeloopColorTheme.sakuraPink);
    expect(sanitized.backgroundPattern, MeloopBackgroundPattern.tealRadial);
    expect(sanitized.stickers.length, 1);
  });

  testWidgets(
    'viewing another user profile temporarily uses default theme and restores own theme on close',
    (tester) async {
      tester.view.physicalSize = const Size(1600, 1000);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final themeCtrl = MeloopThemeController.instance;
      themeCtrl.updateConfig(
        const ProfileCustomizationConfig(
          appStyle: MeloopAppStyle.neonNight,
          colorTheme: MeloopColorTheme.sakuraPink,
        ),
      );
      expect(themeCtrl.colorTheme, MeloopColorTheme.sakuraPink);
      expect(themeCtrl.appStyle, MeloopAppStyle.neonNight);

      await tester.pumpWidget(
        const MaterialApp(
          home: ProfileScreen(username: 'Vicente', insideDesktopShell: true),
        ),
      );
      await tester.pumpAndSettle();

      // Otro perfil muestra el tema por defecto de Meloop sin pisar el propio
      expect(themeCtrl.isViewingOtherProfile, isTrue);
      expect(themeCtrl.colorTheme, MeloopColorTheme.meloopTeal);
      expect(themeCtrl.appStyle, MeloopAppStyle.paper);
      expect(themeCtrl.ownConfig.colorTheme, MeloopColorTheme.sakuraPink);

      // Al salir del perfil ajeno se restaura el tema personalizado del usuario
      await tester.pumpWidget(const MaterialApp(home: SizedBox.shrink()));
      await tester.pumpAndSettle();
      expect(themeCtrl.isViewingOtherProfile, isFalse);
      expect(themeCtrl.colorTheme, MeloopColorTheme.sakuraPink);
      expect(themeCtrl.appStyle, MeloopAppStyle.neonNight);
    },
  );
}

class _ProfilePostsAdapter implements HttpClientAdapter {
  _ProfilePostsAdapter(this.respond);

  final ResponseBody Function(RequestOptions options) respond;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return respond(options);
  }

  @override
  void close({bool force = false}) {}
}
