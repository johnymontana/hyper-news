"""
Geocoding module that uses Ollama to convert a location string to latitude and longitude coordinates.
Returns the result as GeoJSON.
"""

import json
import requests
from typing import Dict, Any, Optional, Tuple


def geocode_location(location: str, model: str = "llama3") -> Dict[str, Any]:
    """
    Geocode a location string to latitude and longitude using Ollama.
    
    Args:
        location: A string describing a location (e.g., "New York City", "Paris, France")
        model: The Ollama model to use (default: "llama3")
        
    Returns:
        A GeoJSON dictionary with a single point feature containing the latitude and longitude
    """
    # Get coordinates from Ollama
    lat, lon = get_coordinates_from_ollama(location, model)
    
    # Create GeoJSON object
    geojson = {
        "type": "Point",
        "coordinates": [lon, lat]  # GeoJSON uses [longitude, latitude] order
    }
    
    return geojson


def get_coordinates_from_ollama(location: str, model: str = "llama3.2") -> Tuple[float, float]:
    """
    Query Ollama to get the latitude and longitude of a location.
    
    Args:
        location: A string describing a location
        model: The Ollama model to use
        
    Returns:
        A tuple of (latitude, longitude)
    """
    prompt = f"""
    I need the precise latitude and longitude coordinates for this location: {location}
    
    Please respond with ONLY a valid JSON object in this exact format:
    {{
        "latitude": [latitude as a float],
        "longitude": [longitude as a float]
    }}
    
    Don't include any explanation or additional text, just the JSON object.
    """
    
    try:
        # Call Ollama API
        response = requests.post(
            "http://localhost:11434/api/generate",
            json={
                "model": "llama3.2",
                "prompt": prompt,
                "stream": False
            },
            timeout=30
        )
        
        response.raise_for_status()
        result = response.json()
        
        # Extract the JSON response from the generation
        # We need to parse the response text to extract just the JSON part
        response_text = result.get("response", "")
        
        # Find JSON object in the response
        json_start = response_text.find("{")
        json_end = response_text.rfind("}") + 1
        
        if json_start >= 0 and json_end > json_start:
            json_str = response_text[json_start:json_end]
            coordinates = json.loads(json_str)
            
            return float(coordinates["latitude"]), float(coordinates["longitude"])
        else:
            raise ValueError("Could not extract valid JSON coordinates from the model response")
            
    except (requests.RequestException, json.JSONDecodeError, ValueError, KeyError) as e:
        raise Exception(f"Error geocoding location '{location}': {str(e)}")


def geocode_to_string(location: str, model: str = "llama3") -> str:
    """
    Geocode a location and return the GeoJSON as a string.
    
    Args:
        location: A string describing a location
        model: The Ollama model to use
        
    Returns:
        A GeoJSON string
    """
    geojson = geocode_location(location, model)
    return json.dumps(geojson, indent=2)


if __name__ == "__main__":
    import sys
    
    if len(sys.argv) > 1:
        location = sys.argv[1]
        try:
            result = geocode_to_string(location)
            print(result)
        except Exception as e:
            print(f"Error: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        print("Usage: python geocode_ollama.py 'location name'", file=sys.stderr)
        sys.exit(1)
