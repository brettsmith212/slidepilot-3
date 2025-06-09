#!/usr/bin/env python3
import uno
import sys
import os
import json
import re
from com.sun.star.connection import NoConnectException

def parse_bullet_points(text):
    """Parse bullet points from text content"""
    if not text or not text.strip():
        return []
    
    # Split by common bullet point patterns
    lines = text.split('\n')
    bullet_points = []
    
    for i, line in enumerate(lines):
        line = line.strip()
        if line:
            # Remove common bullet markers
            clean_text = re.sub(r'^[•·*-]\s*', '', line)
            if clean_text:
                bullet_points.append({
                    "index": i,
                    "text": clean_text
                })
    
    return bullet_points

def get_shape_type(shape):
    """Determine the type of shape based on its properties and content"""
    try:
        # Check if it has text
        if not hasattr(shape, 'getString'):
            return "non_text"
        
        text = shape.getString().strip()
        if not text:
            return "empty_text"
        
        # Try to determine based on shape service name or type
        if hasattr(shape, 'getShapeType'):
            shape_type = shape.getShapeType()
            if 'text' in shape_type.lower():
                # Check position and content to guess if it's title, subtitle, or content
                if len(text) < 100 and '\n' not in text:
                    return "title"
                elif '•' in text or '*' in text or '-' in text:
                    return "content"
                else:
                    return "text_box"
        
        # Fallback logic based on content
        if len(text) < 50 and '\n' not in text:
            return "title"
        elif '•' in text or '*' in text or text.count('\n') > 1:
            return "content"
        else:
            return "text_box"
            
    except Exception:
        return "unknown"

def read_slide(pptx_path, slide_number):
    """Read detailed content from a specific slide"""
    try:
        # Connect to LibreOffice
        local_context = uno.getComponentContext()
        resolver = local_context.ServiceManager.createInstanceWithContext(
            "com.sun.star.bridge.UnoUrlResolver", local_context)
        
        # Connect to the running LibreOffice instance
        context = resolver.resolve("uno:socket,host=localhost,port=8100;urp;StarOffice.ComponentContext")
        desktop = context.ServiceManager.createInstanceWithContext(
            "com.sun.star.frame.Desktop", context)
        
        # Convert file path to file URL
        file_url = uno.systemPathToFileUrl(os.path.abspath(pptx_path))
        
        # Load the presentation with hidden properties
        from com.sun.star.beans import PropertyValue
        
        props = (
            PropertyValue("Hidden", 0, True, 0),
            PropertyValue("ReadOnly", 0, True, 0),
        )
        
        doc = desktop.loadComponentFromURL(file_url, "_blank", 0, props)
        
        # Get the slides
        slides = doc.getDrawPages()
        slide_count = slides.getCount()
        
        # Validate slide number (convert from 1-based to 0-based)
        slide_index = slide_number - 1
        if slide_index < 0 or slide_index >= slide_count:
            raise ValueError(f"Slide number {slide_number} out of range (1-{slide_count})")
        
        # Get the specific slide
        slide = slides.getByIndex(slide_index)
        
        slide_info = {
            "slide_number": slide_number,
            "total_shapes": slide.getCount(),
            "shapes": []
        }
        
        # Extract information from each shape
        for shape_index in range(slide.getCount()):
            shape = slide.getByIndex(shape_index)
            
            shape_info = {
                "shape_index": shape_index,
                "shape_type": get_shape_type(shape),
                "text": "",
                "description": ""
            }
            
            # Try to get text content
            if hasattr(shape, 'getString'):
                text = shape.getString().strip()
                shape_info["text"] = text
                
                # Generate description
                if text:
                    if shape_info["shape_type"] == "title":
                        shape_info["description"] = "Main slide title"
                    elif shape_info["shape_type"] == "content":
                        shape_info["description"] = "Content text box with bullet points"
                        # Parse bullet points
                        bullet_points = parse_bullet_points(text)
                        if bullet_points:
                            shape_info["bullet_points"] = bullet_points
                    else:
                        shape_info["description"] = f"Text box containing: {text[:50]}..."
                else:
                    shape_info["description"] = "Empty text shape"
            else:
                shape_info["description"] = "Non-text shape (image, chart, etc.)"
            
            slide_info["shapes"].append(shape_info)
        
        # Close the document
        doc.close(True)
        
        return slide_info
        
    except NoConnectException:
        raise Exception("Could not connect to LibreOffice. Make sure it's running with UNO socket.")
    except Exception as e:
        raise Exception(f"Error reading slide: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python3 uno_read_slide.py <pptx_path> <slide_number>")
        sys.exit(1)
    
    pptx_path = sys.argv[1]
    slide_number = int(sys.argv[2])
    
    try:
        result = read_slide(pptx_path, slide_number)
        print(json.dumps(result, indent=2))
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
