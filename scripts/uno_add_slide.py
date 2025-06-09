#!/usr/bin/env python3
import uno
import sys
import os
import json
from com.sun.star.connection import NoConnectException

def add_slide(pptx_path, position=None, layout="blank", title=None):
    """Add a new slide to a presentation with optional initial content"""
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
        slide_count = slides.getCount()
        
        # Determine position (default to end if not specified)
        if position is None or position > slide_count:
            position = slide_count
        else:
            # Convert to 0-based index and ensure it's valid
            position = max(0, min(position - 1, slide_count))
        
        # Insert new slide at specified position
        new_slide = slides.insertNewByIndex(position)
        
        # Add title if provided
        if title:
            # Create a title text box
            try:
                # Create a text shape for the title
                shape_service = doc.createInstance("com.sun.star.drawing.TextShape")
                
                # Standard PowerPoint slide dimensions (assuming 10 inch wide, 7.5 inch tall)
                # LibreOffice uses 1/100mm units, so 1 inch = 2540 units
                slide_width = 25400  # 10 inches
                slide_height = 19050  # 7.5 inches
                
                # Position the title at the top center of the slide
                title_width = int(slide_width * 0.8)  # 80% of slide width
                title_height = int(slide_height * 0.15)  # 15% of slide height
                title_x = int((slide_width - title_width) / 2)  # Center horizontally
                title_y = int(slide_height * 0.1)  # 10% from top
                
                # Set position and size
                from com.sun.star.awt import Point, Size
                shape_service.setPosition(Point(title_x, title_y))
                shape_service.setSize(Size(title_width, title_height))
                
                # Add the shape to the slide
                new_slide.add(shape_service)
                
                # Set the text content
                shape_service.setString(title)
                
                # Optional: Set title formatting
                try:
                    text_cursor = shape_service.createTextCursor()
                    text_cursor.gotoStart(False)
                    text_cursor.gotoEnd(True)
                    text_cursor.setPropertyValue("CharHeight", 24.0)  # Font size
                    text_cursor.setPropertyValue("CharWeight", 150.0)  # Bold
                except:
                    pass  # Formatting is optional, don't fail if it doesn't work
                    
            except Exception as e:
                # If title creation fails, continue without it
                pass
        
        # Save the document
        doc.store()
        
        # Get updated slide count
        new_slide_count = slides.getCount()
        new_slide_number = position + 1  # Convert back to 1-based
        
        # Close the document
        doc.close(True)
        
        return {
            "success": True,
            "new_slide_number": new_slide_number,
            "total_slides": new_slide_count,
            "message": f"Successfully added slide {new_slide_number} of {new_slide_count}",
            "title": title if title else "Untitled"
        }
        
    except NoConnectException:
        raise Exception("Could not connect to LibreOffice. Make sure it's running with UNO socket.")
    except Exception as e:
        raise Exception(f"Error adding slide: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 uno_add_slide.py <pptx_path> [position] [layout] [title]")
        sys.exit(1)
    
    pptx_path = sys.argv[1]
    position = None
    layout = "blank"
    title = None
    
    # Parse optional arguments
    if len(sys.argv) > 2 and sys.argv[2].isdigit():
        position = int(sys.argv[2])
    if len(sys.argv) > 3:
        layout = sys.argv[3]
    if len(sys.argv) > 4:
        title = sys.argv[4]
    
    try:
        result = add_slide(pptx_path, position, layout, title)
        print(json.dumps(result, indent=2))
    except Exception as e:
        error_result = {
            "success": False,
            "error": str(e)
        }
        print(json.dumps(error_result, indent=2))
        sys.exit(1)
