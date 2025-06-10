#!/usr/bin/env python3
import uno
import sys
import os
import json
from com.sun.star.connection import NoConnectException
from com.sun.star.beans import PropertyValue
from com.sun.star.text.WritingMode import LR_TB
from com.sun.star.style.NumberingType import ARABIC
from slide_analyzer import SlideAnalyzer

def format_as_bullet_list(shape, bullet_text):
    """Format text shape as a proper bullet list using LibreOffice UNO API"""
    try:
        # Set the text content
        shape.setString(bullet_text)
        
        # Get the text object for formatting
        if hasattr(shape, 'getText'):
            text_obj = shape.getText()
            cursor = text_obj.createTextCursor()
            cursor.gotoStart(False)
            cursor.gotoEnd(True)  # Select all text
            
            # Create numbering rules for bullet points
            numbering_rules = cursor.getPropertyValue("NumberingRules")
            if numbering_rules is None:
                # Create new numbering rules if none exist
                service_manager = cursor.getPropertyValue("ServiceManager") 
                if service_manager is None:
                    # Fallback: just set the text without special formatting
                    return True
                
            # Set bullet point properties
            try:
                # Set basic paragraph properties for bullet list
                cursor.setPropertyValue("NumberingLevel", 0)
                cursor.setPropertyValue("ParaLeftMargin", 500)  # Left indent
                cursor.setPropertyValue("ParaFirstLineIndent", -300)  # Hanging indent
                return True
            except:
                # If numbering properties fail, just set the text
                return True
                
    except Exception as e:
        print(f"Warning: Could not apply bullet formatting: {e}")
        # Fallback to simple text setting
        shape.setString(bullet_text)
        return True
    
    return True

def edit_slide_text(pptx_path, slide_number, target_type, target_value, new_text, old_text=None):
    """Edit text content on a slide using various targeting methods"""
    try:
        # Convert literal \n to actual newlines in new_text
        new_text = new_text.replace('\\n', '\n')
        
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
        
        # Load the presentation (NOT read-only for editing)
        from com.sun.star.beans import PropertyValue
        
        props = (
            PropertyValue("Hidden", 0, True, 0),
            PropertyValue("ReadOnly", 0, False, 0),  # Allow editing
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
        
        # Track if we made any changes
        changes_made = False
        change_description = ""
        
        if target_type == "shape_index":
            # Edit specific shape by index
            shape_index = int(target_value)
            if shape_index < 0 or shape_index >= slide.getCount():
                raise ValueError(f"Shape index {shape_index} out of range (0-{slide.getCount()-1})")
            
            shape = slide.getByIndex(shape_index)
            if hasattr(shape, 'setString'):
                old_text_actual = shape.getString()
                shape.setString(new_text)
                changes_made = True
                change_description = f"Changed shape {shape_index} from '{old_text_actual}' to '{new_text}'"
            else:
                raise ValueError(f"Shape {shape_index} does not contain editable text")
                
        elif target_type == "shape_type":
            # Edit by shape type using shared analyzer
            target_shape_type = target_value.lower()
            
            for i in range(slide.getCount()):
                shape = slide.getByIndex(i)
                
                # Use shared analyzer to determine shape type
                detected_shape_type = SlideAnalyzer.get_shape_type(shape)
                
                should_edit = False
                if target_shape_type == detected_shape_type:
                    should_edit = True
                elif target_shape_type == "content" and detected_shape_type == "bullet_list":
                    # Backward compatibility: "content" maps to "bullet_list"
                    should_edit = True
                
                if should_edit:
                    old_text_actual = shape.getString() if hasattr(shape, 'getString') else ""
                    shape.setString(new_text)
                    changes_made = True
                    change_description = f"Changed {detected_shape_type} (shape {i}) from '{old_text_actual}' to '{new_text}'"
                    break  # Only edit the first matching shape
            
            if not changes_made:
                raise ValueError(f"No shape of type '{target_shape_type}' found on slide {slide_number}")
                
        elif target_type == "text_replace":
            # Replace specific text across all shapes
            if not old_text:
                raise ValueError("old_text is required for text_replace mode")
            
            for i in range(slide.getCount()):
                shape = slide.getByIndex(i)
                if hasattr(shape, 'getString'):
                    current_text = shape.getString()
                    if old_text in current_text:
                        new_full_text = current_text.replace(old_text, new_text)
                        shape.setString(new_full_text)
                        changes_made = True
                        change_description = f"Replaced '{old_text}' with '{new_text}' in shape {i}"
                        break  # Only replace in first matching shape
            
            if not changes_made:
                raise ValueError(f"Text '{old_text}' not found on slide {slide_number}")
                
        elif target_type == "bullet_point":
            # Edit specific bullet point (more complex, simplified for now)
            bullet_index = int(target_value)
            
            for i in range(slide.getCount()):
                shape = slide.getByIndex(i)
                if hasattr(shape, 'getString'):
                    text = shape.getString()
                    if '•' in text or '*' in text or '\n' in text:
                        lines = text.split('\n')
                        if bullet_index < len(lines):
                            lines[bullet_index] = new_text
                            new_full_text = '\n'.join(lines)
                            shape.setString(new_full_text)
                            changes_made = True
                            change_description = f"Changed bullet point {bullet_index} to '{new_text}' in shape {i}"
                            break
            
            if not changes_made:
                raise ValueError(f"Bullet point {bullet_index} not found on slide {slide_number}")
                
        elif target_type == "bullet_list":
            # Format as a bullet list - target_value is the shape index
            shape_index = int(target_value)
            if shape_index < 0 or shape_index >= slide.getCount():
                raise ValueError(f"Shape index {shape_index} out of range (0-{slide.getCount()-1})")
            
            shape = slide.getByIndex(shape_index)
            if hasattr(shape, 'setString'):
                old_text_actual = shape.getString() if hasattr(shape, 'getString') else ""
                # Use the bullet list formatting function
                format_as_bullet_list(shape, new_text)
                changes_made = True
                change_description = f"Set shape {shape_index} as bullet list: '{new_text[:50]}...'"
            else:
                raise ValueError(f"Shape {shape_index} does not contain editable text")
        else:
            raise ValueError(f"Unknown target_type: {target_type}")
        
        if changes_made:
            # Save the document
            doc.store()
            # Don't print success message here - it interferes with JSON output
        
        # Close the document
        doc.close(True)
        
        return {
            "success": changes_made,
            "message": change_description if changes_made else "No changes made",
            "slide_number": slide_number,
            "target_type": target_type,
            "target_value": target_value
        }
        
    except NoConnectException:
        raise Exception("Could not connect to LibreOffice. Make sure it's running with UNO socket.")
    except Exception as e:
        raise Exception(f"Error editing slide: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 6:
        print("Usage: python3 uno_edit_slide.py <pptx_path> <slide_number> <target_type> <target_value> <new_text> [old_text]")
        print(f"target_type: {SlideAnalyzer.EDIT_TARGET_SHAPE_INDEX}, {SlideAnalyzer.EDIT_TARGET_SHAPE_TYPE}, {SlideAnalyzer.EDIT_TARGET_BULLET_POINT}, {SlideAnalyzer.EDIT_TARGET_BULLET_LIST}, {SlideAnalyzer.EDIT_TARGET_TEXT_REPLACE}")
        print("target_value: index/type/text depending on target_type")
        sys.exit(1)
    
    pptx_path = sys.argv[1]
    slide_number = int(sys.argv[2])
    target_type = sys.argv[3]
    target_value = sys.argv[4]
    new_text = sys.argv[5]
    old_text = sys.argv[6] if len(sys.argv) > 6 else None
    
    try:
        result = edit_slide_text(pptx_path, slide_number, target_type, target_value, new_text, old_text)
        print(json.dumps(result, indent=2))
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
