import 'package:flutter/material.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/features/search/presentation/global_search_route.dart';
import 'package:flutter_app/widgets/desktop_app_header.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets(
    'desktop header uses the three navigation pills and bottom radius',
    (tester) async {
      tester.view.physicalSize = const Size(1280, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final controller = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      )..start();
      addTearDown(controller.dispose);
      var selectedArtists = false;

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
                onSearchTap: () {},
              ),
            ),
          ),
        ),
      );

      expect(find.text('Inicio'), findsOneWidget);
      expect(find.text('Amigos'), findsOneWidget);
      expect(find.text('Artistas y canciones'), findsOneWidget);
      expect(find.text('Perfil'), findsNothing);
      expect(find.byTooltip('Ajustes'), findsOneWidget);

      final selectedButton = find.byWidgetPredicate(
        (widget) =>
            widget is Container &&
            widget.decoration is BoxDecoration &&
            (widget.decoration! as BoxDecoration).gradient != null &&
            (widget.decoration! as BoxDecoration).borderRadius ==
                BorderRadius.circular(24),
      );
      expect(tester.getSize(selectedButton).height, 42);

      final navigationPill = find.byWidgetPredicate(
        (widget) =>
            widget is Container &&
            widget.decoration is BoxDecoration &&
            (widget.decoration! as BoxDecoration).color ==
                const Color(0xFF16877F),
      );
      expect(
        tester.getSize(navigationPill).width,
        tester.getSize(selectedButton).width * 3 + 10,
      );
      final searchButton = find.byTooltip('Búsqueda global');
      expect(tester.getSize(searchButton), const Size(48, 48));
      expect(tester.getCenter(navigationPill).dx, 640);
      expect(
        tester.getTopLeft(searchButton).dx,
        greaterThan(tester.getTopRight(navigationPill).dx),
      );
      expect(
        tester.getTopLeft(searchButton).dx -
            tester.getTopRight(navigationPill).dx,
        12,
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
      expect(header, isNotNull);
      final headerDecoration = header.decoration! as BoxDecoration;
      final headerGradient = headerDecoration.gradient! as LinearGradient;
      expect(headerGradient.colors, const [
        Color(0xFF59E5C9),
        Color(0xFF00AF92),
      ]);
      expect(
        tester
            .widget<Image>(
              find.byWidgetPredicate(
                (widget) =>
                    widget is Image &&
                    widget.image is AssetImage &&
                    (widget.image as AssetImage).assetName ==
                        'assets/imagenes/meloop.png',
              ),
            )
            .height,
        38,
      );

      await tester.tap(find.text('Artistas y canciones'));
      expect(selectedArtists, isTrue);
    },
  );

  testWidgets(
    'desktop section navigation returns to home without stale routes',
    (tester) async {
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

      await tester.tap(find.text('Artistas y canciones'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('section-Artistas')), findsOneWidget);
      final artistsRoute = ModalRoute.of(
        tester.element(find.byKey(const ValueKey('section-Artistas'))),
      )!;
      expect(artistsRoute.transitionDuration, Duration.zero);
      expect(artistsRoute.reverseTransitionDuration, Duration.zero);

      await tester.tap(find.text('Amigos'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('section-Amigos')), findsOneWidget);

      await tester.tap(find.text('Inicio'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('section-Inicio')), findsOneWidget);
      expect(find.byKey(const ValueKey('section-Artistas')), findsNothing);
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
