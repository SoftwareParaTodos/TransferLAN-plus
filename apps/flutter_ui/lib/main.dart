import 'package:flutter/material.dart';
import 'screens/home_screen.dart';

void main() {
  runApp(const TransferLanApp());
}

class TransferLanApp extends StatelessWidget {
  const TransferLanApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'TransferLAN+',
      theme: ThemeData(
        brightness: Brightness.dark,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF38BDF8),
          brightness: Brightness.dark,
        ),
        scaffoldBackgroundColor: const Color(0xFF07111F),
        useMaterial3: true,
      ),
      home: const HomeScreen(),
    );
  }
}
