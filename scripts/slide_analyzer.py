#!/usr/bin/env python3
"""
Shared slide analysis module for LibreOffice UNO operations.

This module provides centralized logic for:
- Shape type detection and classification
- Bullet point parsing and analysis
- Content analysis heuristics
- Slide data structure definitions

Used by both read_slide and edit_slide scripts to ensure consistency.
"""

import re
from typing import List, Dict, Any, Optional
from dataclasses import dataclass


@dataclass
class BulletPoint:
    """Represents a single bullet point in a list."""
    index: int
    text: str
    original_text: str = ""  # Preserves original formatting


@dataclass
class ShapeInfo:
    """Comprehensive information about a slide shape."""
    shape_index: int
    shape_type: str
    text: str
    description: str
    bullet_points: Optional[List[BulletPoint]] = None
    edit_hint: Optional[str] = None
    

class SlideAnalyzer:
    """Centralized logic for analyzing slide shapes and content."""
    
    # Shape type constants
    SHAPE_TYPE_TITLE = "title"
    SHAPE_TYPE_BULLET_LIST = "bullet_list"
    SHAPE_TYPE_TEXT_BOX = "text_box"
    SHAPE_TYPE_NON_TEXT = "non_text"
    SHAPE_TYPE_EMPTY_TEXT = "empty_text"
    SHAPE_TYPE_UNKNOWN = "unknown"
    
    # Edit target type constants (for consistency with edit operations)
    EDIT_TARGET_SHAPE_INDEX = "shape_index"
    EDIT_TARGET_SHAPE_TYPE = "shape_type"
    EDIT_TARGET_BULLET_LIST = "bullet_list"
    EDIT_TARGET_TEXT_REPLACE = "text_replace"
    EDIT_TARGET_BULLET_POINT = "bullet_point"
    
    @staticmethod
    def parse_bullet_points(text: str) -> List[BulletPoint]:
        """Parse bullet points from text content."""
        if not text or not text.strip():
            return []
        
        # Split by lines
        lines = text.split('\n')
        bullet_points = []
        
        for i, line in enumerate(lines):
            original_line = line
            line = line.strip()
            if line:
                # Remove common bullet markers
                clean_text = re.sub(r'^[•·*-]\s*', '', line)
                if clean_text:
                    bullet_points.append(BulletPoint(
                        index=i,
                        text=clean_text,
                        original_text=original_line
                    ))
        
        return bullet_points
    
    @staticmethod
    def clean_text_for_bullet_formatting(text: str) -> str:
        """Clean text by removing bullet characters for LibreOffice bullet formatting.
        
        LibreOffice will add its own bullet characters, so we need to remove any
        existing ones to avoid double bullets (•• instead of •).
        """
        if not text:
            return text
            
        # Split into lines and clean each line
        lines = text.split('\n')
        cleaned_lines = []
        
        for line in lines:
            # Remove leading bullet characters and whitespace
            cleaned_line = re.sub(r'^\s*[•·*-]+\s*', '', line)
            # Keep the line even if it becomes empty (preserves line structure)
            cleaned_lines.append(cleaned_line)
        
        return '\n'.join(cleaned_lines)
    
    @staticmethod
    def is_bullet_content(text: str) -> bool:
        """Determine if text contains bullet point content."""
        if not text:
            return False
            
        # Check for bullet characters
        has_bullets = any(char in text for char in ['•', '·', '*', '-'])
        
        # Check for multiple lines (common in bullet lists)
        has_multiple_lines = text.count('\n') > 1
        
        # Check for bullet patterns at line starts
        lines = text.split('\n')
        bullet_pattern_lines = sum(1 for line in lines 
                                 if re.match(r'^\s*[•·*-]\s+', line))
        has_bullet_patterns = bullet_pattern_lines >= 2
        
        return has_bullets or has_multiple_lines or has_bullet_patterns
    
    @staticmethod
    def is_title_content(text: str) -> bool:
        """Determine if text appears to be a title."""
        if not text:
            return False
            
        # Titles are typically short and single-line
        is_short = len(text) < 100
        is_single_line = '\n' not in text
        has_no_bullets = not SlideAnalyzer.is_bullet_content(text)
        
        return is_short and is_single_line and has_no_bullets
    
    @staticmethod
    def get_shape_type(shape) -> str:
        """Determine the type of shape based on its properties and content."""
        try:
            # Check if it has text
            if not hasattr(shape, 'getString'):
                return SlideAnalyzer.SHAPE_TYPE_NON_TEXT
            
            text = shape.getString().strip()
            if not text:
                return SlideAnalyzer.SHAPE_TYPE_EMPTY_TEXT
            
            # Analyze content to determine type
            if SlideAnalyzer.is_title_content(text):
                return SlideAnalyzer.SHAPE_TYPE_TITLE
            elif SlideAnalyzer.is_bullet_content(text):
                return SlideAnalyzer.SHAPE_TYPE_BULLET_LIST
            else:
                return SlideAnalyzer.SHAPE_TYPE_TEXT_BOX
                
        except Exception:
            return SlideAnalyzer.SHAPE_TYPE_UNKNOWN
    
    @staticmethod
    def analyze_shape(shape, shape_index: int) -> ShapeInfo:
        """Perform complete analysis of a shape and return structured info."""
        shape_type = SlideAnalyzer.get_shape_type(shape)
        text = ""
        description = ""
        bullet_points = None
        edit_hint = None
        
        # Extract text content
        if hasattr(shape, 'getString'):
            text = shape.getString().strip()
        
        # Generate description and additional info based on type
        if not text:
            if shape_type == SlideAnalyzer.SHAPE_TYPE_NON_TEXT:
                description = "Non-text shape (image, chart, etc.)"
            else:
                description = "Empty text shape"
        else:
            if shape_type == SlideAnalyzer.SHAPE_TYPE_TITLE:
                description = "Main slide title"
                edit_hint = f"Use target_type='{SlideAnalyzer.EDIT_TARGET_SHAPE_INDEX}' with target_value='{shape_index}' to edit"
                
            elif shape_type == SlideAnalyzer.SHAPE_TYPE_BULLET_LIST:
                description = "Bullet list content"
                bullet_points = SlideAnalyzer.parse_bullet_points(text)
                edit_hint = f"Use target_type='{SlideAnalyzer.EDIT_TARGET_BULLET_LIST}' with target_value='{shape_index}' for proper bullet formatting. Provide text WITHOUT bullet characters - LibreOffice will add them automatically."
                
            elif shape_type == SlideAnalyzer.SHAPE_TYPE_TEXT_BOX:
                description = f"Text box containing: {text[:50]}..."
                edit_hint = f"Use target_type='{SlideAnalyzer.EDIT_TARGET_SHAPE_INDEX}' with target_value='{shape_index}' to edit"
                
            else:
                description = f"Unknown shape containing: {text[:50]}..."
        
        return ShapeInfo(
            shape_index=shape_index,
            shape_type=shape_type,
            text=text,
            description=description,
            bullet_points=bullet_points,
            edit_hint=edit_hint
        )
    
    @staticmethod
    def should_target_as_bullet_list(shape_info: ShapeInfo) -> bool:
        """Determine if a shape should be edited using bullet_list target type."""
        return shape_info.shape_type == SlideAnalyzer.SHAPE_TYPE_BULLET_LIST
    
    @staticmethod
    def get_edit_target_recommendations(shape_info: ShapeInfo) -> Dict[str, str]:
        """Get recommended edit target types for a shape."""
        recommendations = {}
        
        if shape_info.shape_type == SlideAnalyzer.SHAPE_TYPE_BULLET_LIST:
            recommendations["bullet_formatting"] = f"target_type='{SlideAnalyzer.EDIT_TARGET_BULLET_LIST}', target_value='{shape_info.shape_index}'"
            recommendations["individual_bullets"] = f"target_type='{SlideAnalyzer.EDIT_TARGET_BULLET_POINT}', target_value='<bullet_index>'"
            
        if shape_info.text:
            recommendations["replace_all_text"] = f"target_type='{SlideAnalyzer.EDIT_TARGET_SHAPE_INDEX}', target_value='{shape_info.shape_index}'"
            recommendations["replace_specific_text"] = f"target_type='{SlideAnalyzer.EDIT_TARGET_TEXT_REPLACE}', old_text='<text_to_find>'"
            
        return recommendations


def convert_shape_info_to_dict(shape_info: ShapeInfo) -> Dict[str, Any]:
    """Convert ShapeInfo dataclass to dictionary for JSON serialization."""
    result = {
        "shape_index": shape_info.shape_index,
        "shape_type": shape_info.shape_type,
        "text": shape_info.text,
        "description": shape_info.description
    }
    
    if shape_info.bullet_points:
        result["bullet_points"] = [
            {
                "index": bp.index,
                "text": bp.text
            }
            for bp in shape_info.bullet_points
        ]
    
    if shape_info.edit_hint:
        result["edit_hint"] = shape_info.edit_hint
        
    return result
