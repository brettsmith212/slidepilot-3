#!/usr/bin/env python3
import uno
import sys
import os
import json
from com.sun.star.connection import NoConnectException
from slide_analyzer import SlideAnalyzer, convert_shape_info_to_dict

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
        
        # Extract information from each shape using the shared analyzer
        for shape_index in range(slide.getCount()):
            shape = slide.getByIndex(shape_index)
            
            # Use the shared analyzer for consistent shape analysis
            shape_info = SlideAnalyzer.analyze_shape(shape, shape_index)
            
            # Convert to dictionary format for JSON output
            shape_dict = convert_shape_info_to_dict(shape_info)
            
            slide_info["shapes"].append(shape_dict)
        
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
