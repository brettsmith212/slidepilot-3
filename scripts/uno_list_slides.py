#!/usr/bin/env python3
import uno
import sys
import os
import json
from com.sun.star.connection import NoConnectException

def list_slides(pptx_path):
    """List all slides in a presentation with basic information"""
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
        
        slides_info = []
        
        for i in range(slide_count):
            slide = slides.getByIndex(i)
            slide_info = {
                "slide_number": i + 1,
                "title": "Untitled",
                "layout": "Unknown Layout",
                "text_shapes": 0
            }
            
            # Try to get title and count text shapes
            text_shape_count = 0
            for j in range(slide.getCount()):
                shape = slide.getByIndex(j)
                
                if hasattr(shape, 'getString'):
                    text_shape_count += 1
                    # Try to get title from first text shape
                    if j == 0 and slide_info["title"] == "Untitled":
                        text = shape.getString().strip()
                        if text:
                            # Limit title length for display
                            slide_info["title"] = text[:50] + "..." if len(text) > 50 else text
            
            slide_info["text_shapes"] = text_shape_count
            slides_info.append(slide_info)
        
        # Close the document
        doc.close(True)
        
        return {
            "total_slides": slide_count,
            "slides": slides_info
        }
        
    except NoConnectException:
        raise Exception("Could not connect to LibreOffice. Make sure it's running with UNO socket.")
    except Exception as e:
        raise Exception(f"Error listing slides: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 uno_list_slides.py <pptx_path>")
        sys.exit(1)
    
    pptx_path = sys.argv[1]
    
    try:
        result = list_slides(pptx_path)
        print(json.dumps(result, indent=2))
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
