#!/usr/bin/env python3
"""
Simple test script for the slide_analyzer module.
"""

from slide_analyzer import SlideAnalyzer

def test_bullet_detection():
    """Test bullet point detection logic."""
    print("Testing bullet point detection:")
    
    test_cases = [
        ("Simple title", False),
        ("• First point\n• Second point", True),
        ("* Item one\n* Item two\n* Item three", True),
        ("Multi-line\ncontent\nwithout bullets", True),
        ("", False),
        ("A single line without bullets", False),
        ("- Dash bullets\n- Are detected", True),
    ]
    
    for text, expected in test_cases:
        result = SlideAnalyzer.is_bullet_content(text)
        status = "✓" if result == expected else "✗"
        print(f"  {status} '{text[:30]}...' -> {result} (expected {expected})")

def test_title_detection():
    """Test title detection logic."""
    print("\nTesting title detection:")
    
    test_cases = [
        ("Simple Title", True),
        ("A longer title but still under 100 characters", True),
        ("• This has bullets", False),
        ("Multi-line\ntitle", False),
        ("", False),
        ("A" * 150, False),  # Too long
    ]
    
    for text, expected in test_cases:
        result = SlideAnalyzer.is_title_content(text)
        status = "✓" if result == expected else "✗"
        print(f"  {status} '{text[:30]}...' -> {result} (expected {expected})")

def test_bullet_parsing():
    """Test bullet point parsing."""
    print("\nTesting bullet point parsing:")
    
    text = "• First bullet point\n• Second bullet point\n• Third bullet point"
    bullets = SlideAnalyzer.parse_bullet_points(text)
    
    print(f"  Input: '{text}'")
    print(f"  Parsed {len(bullets)} bullet points:")
    for bp in bullets:
        print(f"    [{bp.index}] {bp.text}")

if __name__ == "__main__":
    print("=== Slide Analyzer Tests ===")
    test_bullet_detection()
    test_title_detection()
    test_bullet_parsing()
    print("\n=== Tests Complete ===")
