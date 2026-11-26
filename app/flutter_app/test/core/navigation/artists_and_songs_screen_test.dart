import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:flutter_app/core/navigation/artists_and_songs_screen.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets(
    'desktop layout shows 5-column album grid, Recomendados/Tendencia tabs, multi-field search, hover vinyl, and friends artists sidebar',
    (tester) async {
      tester.view.physicalSize = const Size(1600, 960);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(const MaterialApp(home: ArtistsAndSongsScreen()));
      await tester.pumpAndSettle();

      // Ambas hojas visibles (izquierda álbumes, derecha artistas de amigos)
      expect(find.text('Artistas y canciones'), findsWidgets);
      expect(find.text('Lo que escuchan tus amigos'), findsOneWidget);
      expect(find.text('Recomendados'), findsOneWidget);
      expect(find.text('Tendencia'), findsOneWidget);

      // Álbum inicial con placeholder de símbolo de imagen
      expect(find.text('Graduation'), findsWidgets);
      expect(find.byIcon(Icons.image_outlined), findsWidgets);

      // Hover sobre un álbum hace aparecer el vinilo a su derecha
      final gesture = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await gesture.addPointer(location: Offset.zero);
      addTearDown(gesture.removePointer);

      await gesture.moveTo(tester.getCenter(find.text('Graduation').last));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey<String>('vinyl-alb-1')), findsOneWidget);

      // Abrir detalle ampliado estilo Spotify y probar Me gusta + Favorito
      await tester.tap(find.text('Graduation').last);
      await tester.pumpAndSettle();
      expect(find.text('ÁLBUM'), findsOneWidget);
      expect(find.text('TÍTULO'), findsOneWidget);
      expect(find.text('Flashing Lights'), findsWidgets);

      await tester.tap(find.text('Me gusta'));
      await tester.pumpAndSettle();
      expect(find.text('Te gusta'), findsOneWidget);

      await tester.tap(find.text('Marcar como favorito'));
      await tester.pumpAndSettle();
      expect(find.text('En tus favoritos'), findsOneWidget);

      await tester.tap(find.byTooltip('Cerrar'));
      await tester.pumpAndSettle();
      expect(find.text('ÁLBUM'), findsNothing);

      // Dejamos expirar los toasts para evitar timers pendientes
      await tester.pump(const Duration(seconds: 3));
      await tester.pumpAndSettle();

      // Cambiar entre Recomendados y Tendencia
      await tester.tap(find.text('Tendencia'));
      await tester.pumpAndSettle();
      expect(find.text('Satirología Vol. 2'), findsOneWidget);

      // Búsqueda por nombre de canción ("Flashing Lights") encuentra el álbum Graduation
      await tester.enterText(find.byType(TextField), 'Flashing Lights');
      await tester.pumpAndSettle();
      expect(find.text('Graduation'), findsWidgets);
      expect(find.text('♪ Flashing Lights'), findsOneWidget);

      // Limpiar búsqueda y filtrar artistas de la derecha por amigo ("Andrea")
      await tester.tap(find.byTooltip('Limpiar búsqueda'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Andrea'));
      await tester.pumpAndSettle();
      expect(find.text('Andrea escucha esto'), findsOneWidget);
      expect(find.text('Felipe escucha esto'), findsNothing);
    },
  );

  testWidgets('mobile layout renders 3 pill tabs and switches to De amigos', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(412, 915);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(const MaterialApp(home: ArtistsAndSongsScreen()));
    await tester.pumpAndSettle();

    expect(find.text('Recomendados'), findsOneWidget);
    expect(find.text('Tendencia'), findsOneWidget);
    expect(find.text('De amigos'), findsOneWidget);

    await tester.tap(find.text('De amigos'));
    await tester.pumpAndSettle();
    expect(find.text('Skrillex'), findsOneWidget);
  });
}
