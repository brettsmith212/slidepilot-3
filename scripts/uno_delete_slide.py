#!/usr/bin/env python3
import uno
import sys
import os
import json
from com.sun.star.connection import NoConnectException

def delete_slide(pptx_path, slide_number):
    """Delete a specific slide from a presentation"""
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
        
        # Load the presentation (not read-only since we're editing)
        from com.sun.star.beans import PropertyValue
        
        props = (
            PropertyValue("Hidden", 0, True, 0),
        )
        
        doc = desktop.loadComponentFromURL(file_url, "_blank", 0, props)
        
        # Get the slides collection
        slides = doc.getDrawPages()
        original_slide_count = slides.getCount()
        
        # Validate slide number
        if slide_number < 1 or slide_number > original_slide_count:
            raise Exception(f"Invalid slide number {slide_number}. Presentation has {original_slide_count} slides.")
        
        # Get slide title before deletion for confirmation
        slide_to_delete = slides.getByIndex(slide_number - 1)  # Convert to 0-based
        deleted_slide_title = "Untitled"
        
        try:
            # Try to get the title from the first text shape
            for j in range(slide_to_delete.getCount()):
                shape = slide_to_delete.getByIndex(j)
                if hasattr(shape, 'getString'):
                    text = shape.getString().strip()
                    if text:
                        deleted_slide_title = text[:50] + "..." if len(text) > 50 else text
                        break
        except:
            pass  # If we can't get the title, that's okay
        
        # Delete the slide
        slides.remove(slide_to_delete)
        
        # Save the document
        doc.store()
        
        # Get updated slide count
        new_slide_count = slides.getCount()
        
        # Close the document
        doc.close(True)
        
        return {
            "success": True,
            "deleted_slide_number": slide_number,
            "deleted_slide_title": deleted_slide_title,
            "original_slide_count": original_slide_count,
            "new_slide_count": new_slide_count,
            "message": f"Successfully deleted slide {slide_number} ('{deleted_slide_title}'). Presentation now has {new_slide_count} slides."
        }
        
    except NoConnectException:
        raise Exception("Could not connect to LibreOffice. Make sure it's running with UNO socket.")
    except Exception as e:
        raise Exception(f"Error deleting slide: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python3 uno_delete_slide.py <pptx_path> <slide_number>")
        sys.exit(1)
    
    pptx_path = sys.argv[1]
    
    try:
        slide_number = int(sys.argv[2])
    except ValueError:
        error_result = {
            "success": False,
            "error": "Slide number must be an integer"
        }
        print(json.dumps(error_result, indent=2))
        sys.exit(1)
    
    try:
        result = delete_slide(pptx_path, slide_number)
        print(json.dumps(result, indent=2))
    except Exception as e:
        error_result = {
            "success": False,
            "error": str(e)
        }
        print(json.dumps(error_result, indent=2))
        sys.exit(1)
