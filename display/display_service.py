"""
This module provides the DisplayService class for managing display operations.
"""

import platform

from PIL import Image


class DisplayService:
    """
    This class manages display operations for the application.
    """

    def __init__(self):
        print("DisplayService initialized")

        # Inky Impression 7.3" has a 800x480 screen resolution
        self.width = 800
        self.height = 480
        self.is_pi = platform.system() == "Linux" and platform.machine().startswith("aarch")

        if self.is_pi:
            from inky.auto import auto
            self.display = auto()
        else:
            print("Not running on Raspberry Pi - display updates will be simulated")
            self.display = None

    async def update_display(self, image_path):
        """
        Updates the display with the image at the specified path.

        Args:
            image_path (str): The path to the image file to be displayed.
        """
        image = Image.open(image_path)

        # Ensure image is correct size
        if image.size != (self.width, self.height):
            print(f"Image size: {image.size}")
            print(f"Resizing image to {self.width}x{self.height}")
            image = image.resize((self.width, self.height))
            image.save(image_path)

        # Display on Inky if available, otherwise save for preview
        if self.display:
            self.display.set_image(image)
            self.display.show()
        else:
            print(f"Image saved as {image_path}")