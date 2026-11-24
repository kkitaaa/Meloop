import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/core/navigation/desktop_navigation_shell.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/features/search/presentation/global_search_route.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/desktop_app_header.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets(
    'desktop header renders navigation pills, bottom radius, and pfp dropdown menu',
    (tester) async {
      tester.view.physicalSize = const Size(1920, 1080);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final controller = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      )..start();
      addTearDown(controller.dispose);
      var selectedArtists = false;
      var openedProfile = false;

      await tester.pumpWidget(
        NotificationScope(
          controller: controller,
          child: MaterialApp(
            home: Scaffold(
              body: DesktopAppHeader(
                activeSection: DesktopNavigationSection.home,
                onHomeTap: () {},
                onFriendsTap: () {},
                onArtistsTap: () => selectedArtists = true,
                onProfileTap: () => openedProfile = true,
                onSearchTap: () {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Inicio'), findsOneWidget);
      expect(find.text('Amigos'), findsOneWidget);
      expect(find.text('Artistas y canciones'), findsOneWidget);
      expect(find.byTooltip('Ajustes'), findsNothing);
      expect(find.byTooltip('Búsqueda global'), findsOneWidget);
      expect(find.byTooltip('Perfil y opciones'), findsOneWidget);

      // Todos los botones de la navbar tienen el mismo ancho y están centrados
      final homeWidth = tester
          .getSize(
            find.ancestor(
              of: find.text('Inicio'),
              matching: find.byType(AnimatedContainer),
            ),
          )
          .width;
      final friendsWidth = tester
          .getSize(
            find.ancestor(
              of: find.text('Amigos'),
              matching: find.byType(AnimatedContainer),
            ),
          )
          .width;
      final artistsWidth = tester
          .getSize(
            find.ancestor(
              of: find.text('Artistas y canciones'),
              matching: find.byType(AnimatedContainer),
            ),
          )
          .width;
      expect(homeWidth, friendsWidth);
      expect(friendsWidth, artistsWidth);

      final pillCenter = tester.getCenter(
        find.byKey(const ValueKey('nav-bar-pill-row')),
      );
      expect(pillCenter.dx, closeTo(960, 1));
      expect(pillCenter.dy, closeTo(42, 0.5));
      expect(
        tester.getCenter(find.byTooltip('Búsqueda global')).dy,
        closeTo(42, 0.5),
      );
      expect(
        tester.getCenter(find.byTooltip('Perfil y opciones')).dy,
        closeTo(42, 0.5),
      );

      final header = tester.widget<Container>(
        find
            .descendant(
              of: find.byType(DesktopAppHeader),
              matching: find.byWidgetPredicate(
                (widget) =>
                    widget is Container &&
                    widget.decoration is BoxDecoration &&
                    (widget.decoration! as BoxDecoration).borderRadius ==
                        const BorderRadius.vertical(
                          bottom: Radius.circular(24),
                        ),
              ),
            )
            .first,
      );
      final headerGradient =
          (header.decoration! as BoxDecoration).gradient! as LinearGradient;
      expect(headerGradient.colors, const [
        Color(0xFF59E5C9),
        Color(0xFF00AF92),
      ]);

      await tester.tap(find.text('Artistas y canciones'));
      expect(selectedArtists, isTrue);

      await tester.tap(find.byTooltip('Perfil y opciones'));
      await tester.pumpAndSettle();
      expect(find.text('Mirar Perfil'), findsOneWidget);
      expect(find.text('Preferencias musicales'), findsOneWidget);
      expect(find.text('Usuarios bloqueados'), findsOneWidget);

      await tester.tap(find.text('Mirar Perfil'));
      await tester.pumpAndSettle();
      expect(openedProfile, isTrue);
    },
  );

  testWidgets(
    'DesktopNavigationShell keeps a single global navbar across sections and profile',
    (tester) async {
      tester.view.physicalSize = const Size(1920, 1080);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final controller = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      )..start();
      addTearDown(controller.dispose);

      final apiClient = ApiClient(
        httpClient: MockClient(
          (_) async => http.Response(
            jsonEncode({
              'success': true,
              'data': {
                'items': [],
                'page': 1,
                'pageSize': 5,
                'hasMore': false,
              },
              'error': null,
            }),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          ),
        ),
      );
      addTearDown(apiClient.close);

      await tester.pumpWidget(
        NotificationScope(
          controller: controller,
          child: MaterialApp(
            home: DesktopNavigationShell(apiClient: apiClient),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byType(DesktopAppHeader), findsOneWidget);
      expect(find.text('Chat'), findsNothing);
      final initialTopLeft = tester.getTopLeft(find.byType(DesktopAppHeader));

      await tester.tap(find.text('Amigos').first);
      await tester.pumpAndSettle();
      expect(find.byType(DesktopAppHeader), findsOneWidget);
      expect(tester.getTopLeft(find.byType(DesktopAppHeader)), initialTopLeft);

      await tester.tap(find.text('Artistas y canciones'));
      await tester.pumpAndSettle();
      expect(find.byType(DesktopAppHeader), findsOneWidget);
      expect(tester.getTopLeft(find.byType(DesktopAppHeader)), initialTopLeft);

      await tester.tap(find.byTooltip('Perfil y opciones'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Mirar Perfil'));
      await tester.pumpAndSettle();
      expect(find.byType(DesktopAppHeader), findsOneWidget);
      expect(tester.getTopLeft(find.byType(DesktopAppHeader)), initialTopLeft);
    },
  );

  testWidgets('desktop search action opens a search field and back returns', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final controller = NotificationController(
      repository: MockNotificationRepository(),
      playSound: false,
    )..start();
    addTearDown(controller.dispose);

    await tester.pumpWidget(
      NotificationScope(
        controller: controller,
        child: const MaterialApp(home: _NavigationTestPage('Inicio')),
      ),
    );

    await tester.tap(find.byTooltip('Búsqueda global'));
    await tester.pumpAndSettle();
    expect(find.byType(TextField), findsOneWidget);

    await tester.tap(find.byTooltip('Volver'));
    await tester.pump(const Duration(milliseconds: 140));
    expect(find.text('Inicio'), findsWidgets);
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('section-Inicio')), findsOneWidget);
  });
}

class _NavigationTestPage extends StatelessWidget {
  const _NavigationTestPage(this.section);

  final String section;

  @override
  Widget build(BuildContext context) {
    final currentSection = switch (section) {
      'Inicio' => DesktopNavigationSection.home,
      'Amigos' => DesktopNavigationSection.friends,
      'Artistas' => DesktopNavigationSection.artists,
      _ => DesktopNavigationSection.none,
    };

    return Scaffold(
      body: Column(
        children: [
          DesktopAppHeader(
            activeSection: currentSection,
            onHomeTap: () => navigateToDesktopHome(context),
            onFriendsTap: () => navigateToDesktopSection(
              context,
              (_) => const _NavigationTestPage('Amigos'),
            ),
            onArtistsTap: () => navigateToDesktopSection(
              context,
              (_) => const _NavigationTestPage('Artistas'),
            ),
            onSearchTap: () => openGlobalSearch(context),
          ),
          Text(section, key: ValueKey('section-$section')),
        ],
      ),
    );
  }
}
