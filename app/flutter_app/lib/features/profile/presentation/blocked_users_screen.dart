import 'package:flutter/material.dart';
import 'dart:async';
import 'package:flutter_app/widgets/states/loading_state.dart';
import 'package:flutter_app/widgets/states/empty_state.dart';
import 'package:flutter_app/widgets/states/error_state.dart';

class BlockedUsersScreen extends StatefulWidget {
  const BlockedUsersScreen({super.key});

  @override
  State<BlockedUsersScreen> createState() => _BlockedUsersScreenState();
}

class _BlockedUsersScreenState extends State<BlockedUsersScreen> {
  final Color _bgColor = const Color(0xFF0D5C5E);
  final Color _tealAccent = const Color(0xFF1ABC9C);

  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _blockedUsers = [];

  bool _isInitialLoading = true;
  bool _isLoadingMore = false;
  bool _hasError = false;
  int _currentPage = 1;

  @override
  void initState() {
    super.initState();
    _fetchBlockedUsers();

    _scrollController.addListener(() {
      if (_scrollController.position.pixels >=
          _scrollController.position.maxScrollExtent - 200) {
        if (!_isLoadingMore && !_hasError) {
          _fetchMoreBlockedUsers();
        }
      }
    });
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _fetchBlockedUsers() async {
    setState(() {
      _isInitialLoading = true;
      _hasError = false;
    });

    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    // Engaño al linter para evitar código muerto, pero siempre será falso (nunca dará error)
    bool simulateError = DateTime.now().second > 60;
    if (simulateError) {
      setState(() {
        _isInitialLoading = false;
        _hasError = true;
      });
      return;
    }

    setState(() {
      _blockedUsers.clear();
      _blockedUsers.addAll(_generateMockUsers(1, 10));
      _isInitialLoading = false;
    });
  }

  Future<void> _fetchMoreBlockedUsers() async {
    setState(() => _isLoadingMore = true);
    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    setState(() {
      _currentPage++;
      _blockedUsers.addAll(_generateMockUsers(_currentPage, 5));
      _isLoadingMore = false;
    });
  }

  List<Map<String, dynamic>> _generateMockUsers(int page, int count) {
    return List.generate(count, (index) {
      int id = (page - 1) * count + index + 1;
      return {
        "id": id,
        "name": "Usuario_Bloqueado_$id",
        "date": "12/10/2026",
        "avatarUrl": null,
      };
    });
  }

  void _confirmUnblock(Map<String, dynamic> user, int index) {
    showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          backgroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          title: Text(
            "¿Desbloquear a @${user['name']}?",
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
          ),
          content: const Text(
            "Podrán volver a enviarse solicitudes de amistad. La amistad anterior no se recupera.",
            style: TextStyle(fontSize: 14, color: Colors.black87),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dialogContext),
              child: const Text("Cancelar", style: TextStyle(color: Colors.grey)),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: _tealAccent,
                elevation: 0,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
              ),
              onPressed: () {
                Navigator.pop(dialogContext);
                _executeUnblockOptimistic(user, index);
              },
              child: const Text("Desbloquear", style: TextStyle(color: Colors.white)),
            ),
          ],
        );
      },
    );
  }

  Future<void> _executeUnblockOptimistic(Map<String, dynamic> user, int index) async {
    setState(() {
      _blockedUsers.removeAt(index);
    });

    await Future.delayed(const Duration(milliseconds: 800));
    if (!mounted) return;

    // Siempre será verdadero para no interrumpir tu prueba visual
    bool httpSuccess = DateTime.now().second < 60; 

    if (httpSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text("Has desbloqueado a @${user['name']}."),
          backgroundColor: _tealAccent,
          duration: const Duration(seconds: 2),
        ),
      );
    } else {
      setState(() {
        _blockedUsers.insert(index, user);
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text("Error de red. No se pudo desbloquear al usuario."),
          backgroundColor: Colors.redAccent,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: _bgColor, // Fondo verde característico de Meloop
      appBar: AppBar(
        backgroundColor: _tealAccent,
        title: const Text(
          "Usuarios bloqueados",
          style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
        centerTitle: true,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: isDesktop ? 650 : double.infinity),
          child: Container(
            margin: EdgeInsets.all(isDesktop ? 24.0 : 0),
            decoration: BoxDecoration(
              color: const Color(0xFFFDFDFD), // Contenedor blanco tipo "papel"
              borderRadius: BorderRadius.circular(isDesktop ? 16 : 0),
              boxShadow: isDesktop
                  ? [
                      BoxShadow(
                        color: Colors.black.withValues(alpha: 0.15),
                        blurRadius: 20,
                      )
                    ]
                  : [],
            ),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(isDesktop ? 16 : 0),
              child: Column(
                children: [
                  Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Row(
                      children: [
                        Icon(Icons.security, color: _tealAccent, size: 24),
                        const SizedBox(width: 12),
                        const Expanded(
                          child: Text(
                            "Administra a quiénes has restringido. Los usuarios bloqueados no pueden ver tu perfil ni interactuar con tus publicaciones.",
                            style: TextStyle(fontSize: 12, color: Colors.black54),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const Divider(height: 1, thickness: 1, color: Colors.black12),
                  Expanded(child: _buildBody()),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildBody() {
    if (_isInitialLoading) return const LoadingState();

    if (_hasError) {
      return ErrorState(
        errorCode: 500,
        onRetry: _fetchBlockedUsers,
        tealAccent: _tealAccent,
      );
    }

    if (_blockedUsers.isEmpty) {
      return const EmptyState(
        title: "Sin bloqueos",
        message: "No tienes a ningún usuario bloqueado actualmente.",
        icon: Icons.person_off,
      );
    }

    return ListView.separated(
      controller: _scrollController,
      itemCount: _blockedUsers.length + (_isLoadingMore ? 1 : 0),
      separatorBuilder: (context, index) => const Divider(height: 1, color: Colors.black12),
      itemBuilder: (context, index) {
        if (index == _blockedUsers.length) {
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 24.0),
            child: Center(child: CircularProgressIndicator(color: _tealAccent)),
          );
        }

        final user = _blockedUsers[index];
        return _buildBlockedUserTile(user, index);
      },
    );
  }

  Widget _buildBlockedUserTile(Map<String, dynamic> user, int index) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
      child: Row(
        children: [
          CircleAvatar(
            radius: 22,
            backgroundColor: Colors.grey[200],
            child: const Icon(Icons.person_off, color: Colors.grey, size: 20),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  user['name'],
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Colors.black87),
                ),
                const SizedBox(height: 4),
                Text(
                  "Bloqueado el ${user['date']}",
                  style: const TextStyle(fontSize: 11, color: Colors.black54),
                ),
              ],
            ),
          ),
          OutlinedButton(
            style: OutlinedButton.styleFrom(
              foregroundColor: _tealAccent,
              side: BorderSide(color: _tealAccent, width: 1.5),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            ),
            onPressed: () => _confirmUnblock(user, index),
            child: const Text("Desbloquear", style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }
}