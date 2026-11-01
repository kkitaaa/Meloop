import 'package:flutter/material.dart';

import 'dart:async';

import 'package:flutter_app/widgets/states/loading_state.dart';
import 'package:flutter_app/widgets/states/empty_state.dart';
import 'package:flutter_app/widgets/states/error_state.dart';
import 'package:flutter_app/features/profile/presentation/profile_screen.dart';

class MutualFriendsScreen extends StatefulWidget {
  final int targetUserId;
  final int initialCount;

  const MutualFriendsScreen({
    super.key,
    required this.targetUserId,
    required this.initialCount,
  });

  @override
  State<MutualFriendsScreen> createState() => _MutualFriendsScreenState();
}

class _MutualFriendsScreenState extends State<MutualFriendsScreen> {
  final Color _tealAccent = const Color(0xFF1ABC9C);
  final ScrollController _scrollController = ScrollController();

  final List _friends = [];

  bool _isInitialLoading = true;
  bool _isLoadingMore = false;

  bool _hasError = false;
  bool _is404 = false;
  bool _isHidden = false;
  int _currentPage = 1;

  @override
  void initState() {
    super.initState();
    _fetchMutualFriends();

    _scrollController.addListener(() {
      if (_scrollController.position.pixels >=
          _scrollController.position.maxScrollExtent - 200) {
        if (!_isLoadingMore && !_hasError && !_is404 && !_isHidden) {
          _fetchMoreFriends();
        }
      }
    });
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _fetchMutualFriends() async {
    setState(() {
      _isInitialLoading = true;
      _hasError = false;
      _is404 = false;
    });

    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    bool simulate404 = DateTime.now().second > 60;
    bool simulateHidden = DateTime.now().second > 60;

    if (simulate404) {
      setState(() {
        _isInitialLoading = false;
        _is404 = true;
      });
      return;
    }

    if (simulateHidden) {
      setState(() {
        _isInitialLoading = false;
        _isHidden = true;
      });
      return;
    }

    setState(() {
      _friends.addAll(_generateMockFriends(1, 15));
      _isInitialLoading = false;
    });
  }

  Future<void> _fetchMoreFriends() async {
    setState(() => _isLoadingMore = true);
    await Future.delayed(const Duration(seconds: 1));
    if (!mounted) return;

    setState(() {
      _currentPage++;
      _friends.addAll(_generateMockFriends(_currentPage, 10));
      _isLoadingMore = false;
    });
  }

  List _generateMockFriends(int page, int count) {
    return List.generate(count, (index) {
      int id = (page - 1) * count + index + 1;
      return {"id": id, "name": "Amigo_Común_$id", "avatarUrl": null};
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: _tealAccent,
        title: const Text(
          "Amigos en común",
          style: TextStyle(color: Colors.white, fontSize: 16),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isInitialLoading) return const LoadingState();

    if (_is404) {
      return const EmptyState(
        title: "Perfil no disponible",
        message: "No se puede acceder a la información de este usuario en este momento.",
        icon: Icons.person_off,
      );
    }

    if (_hasError) {
      return ErrorState(
        errorCode: 500,
        onRetry: _fetchMutualFriends,
        tealAccent: _tealAccent,
      );
    }

    if (_isHidden) {
      return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.visibility_off, size: 64, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            "${widget.initialCount} amigos en común",
            style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 32.0),
            child: Text(
              "Este usuario oculta su lista de amigos.",
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.black54),
            ),
          ),
        ],
      );
    }

    return ListView.builder(
      controller: _scrollController,
      itemCount: _friends.length + (_isLoadingMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (index == _friends.length) {
          return Padding(
            padding: const EdgeInsets.all(16.0),
            child: Center(child: CircularProgressIndicator(color: _tealAccent)),
          );
        }

        final friend = _friends[index] as Map<String, dynamic>;
        return Semantics(
          label: "Ver perfil de ${friend['name']}",
          button: true,
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: Colors.grey[200],
              child: const Icon(Icons.person, color: Colors.grey),
            ),
            title: Text(
              friend['name'],
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) =>
                      const ProfileScreen(username: "Amigo_Común"),
                ),
              );
            },
          ),
        );
      },
    );
  }
}
