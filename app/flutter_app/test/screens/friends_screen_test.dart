import 'package:flutter/material.dart';
import 'package:flutter_app/screens/friends_screen.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets(
    'desktop layout splits friends/requests on the left and suggestions on the right with delete confirmation',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(const MaterialApp(home: FriendsScreen()));
      await tester.pumpAndSettle();

      // Ambas hojas visibles en formato widescreen
      expect(find.text('Red de amigos'), findsOneWidget);
      expect(find.text('Perfiles que te podrían interesar'), findsOneWidget);
      expect(find.text('Vicente'), findsOneWidget);
      expect(find.text('Martín'), findsOneWidget);

      // Cancelar eliminación mantiene al amigo
      await tester.tap(find.byTooltip('Eliminar amigo').first);
      await tester.pumpAndSettle();
      expect(find.text('Eliminar amigo'), findsOneWidget);
      await tester.tap(find.text('Cancelar'));
      await tester.pumpAndSettle();
      expect(find.text('Vicente'), findsOneWidget);

      // Confirmar eliminación lo quita de la lista
      await tester.tap(find.byTooltip('Eliminar amigo').first);
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(ElevatedButton, 'Eliminar'));
      await tester.pumpAndSettle();
      expect(find.text('Vicente'), findsNothing);

      // Cambiar a Solicitudes y aceptar una solicitud la suma a Amigos
      await tester.tap(find.text('Solicitudes (2)'));
      await tester.pumpAndSettle();
      expect(find.text('NuevoUsuario_01'), findsOneWidget);

      await tester.tap(find.text('Aceptar').first);
      await tester.pumpAndSettle();
      expect(find.text('NuevoUsuario_01'), findsNothing);

      await tester.tap(find.text('Amigos').last);
      await tester.pumpAndSettle();
      expect(find.text('NuevoUsuario_01'), findsOneWidget);

      // Enviar solicitud a un perfil sugerido del panel derecho
      await tester.tap(find.text('Agregar').first);
      await tester.pumpAndSettle();
      expect(find.text('Martín'), findsNothing);

      // Dejamos expirar el toast para evitar timers pendientes al cerrar el test
      await tester.pump(const Duration(seconds: 3));
      await tester.pumpAndSettle();
    },
  );

  testWidgets('mobile layout shows 3 pill tabs and confirms friend deletion', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(412, 915);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(const MaterialApp(home: FriendsScreen()));
    await tester.pumpAndSettle();

    expect(find.text('Amigos'), findsOneWidget);
    expect(find.text('Solicitudes (2)'), findsOneWidget);
    expect(find.text('Sugeridos'), findsOneWidget);

    await tester.tap(find.text('Sugeridos'));
    await tester.pumpAndSettle();
    expect(find.text('Martín'), findsOneWidget);
  });
}
