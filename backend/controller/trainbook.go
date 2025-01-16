package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"project-se/entity"
	"project-se/config"
)

// ดึงข้อมูลการจองทั้งหมด
func GetTrainBookings(c *gin.Context) {
	var bookings []entity.TrainBook
	db := config.DB()
	if err := db.Preload("Driver").Preload("Room").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถดึงข้อมูลการจองได้"})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// ดึงข้อมูลการจองตาม ID
func GetTrainBookingByID(c *gin.Context) {
	id := c.Param("id")
	var booking entity.TrainBook
	db := config.DB()
	if err := db.Preload("Driver").Preload("Room").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลการจอง"})
		return
	}
	c.JSON(http.StatusOK, booking)
}

// สร้างการจองใหม่
func CreateTrainBooking(c *gin.Context) {
    var input entity.TrainBook
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
        return
    }

    db := config.DB()

    // ตรวจสอบห้องที่ระบุ
    var room entity.Rooms
    if err := db.First(&room, input.RoomID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลห้อง"})
        return
    }

    // ตรวจสอบผู้ขับที่ระบุ
    var driver entity.Driver
    if err := db.First(&driver, input.DriverID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลผู้ขับ"})
        return
    }

    // ตรวจสอบความจุของห้อง
    if room.CurrentBookings+1 > room.Capacity {
        c.JSON(http.StatusBadRequest, gin.H{"error": "จำนวนการจองเกินความจุของห้อง"})
        return
    }

    // อัปเดตจำนวนการจองในห้อง
    room.CurrentBookings++
    if err := db.Save(&room).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตจำนวนการจองในห้องได้"})
        return
    }

    // บันทึกการจองใหม่
    if err := db.Create(&input).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึกข้อมูลการจองได้"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "สร้างการจองสำเร็จ", "trainbook": input})
}

// อัปเดตการจอง
func UpdateTrainBooking(c *gin.Context) {
    id := c.Param("id")
    var trainbook entity.TrainBook

    db := config.DB()
    if err := db.First(&trainbook, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลการจอง"})
        return
    }

    var input entity.TrainBook
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
        return
    }

    // ตรวจสอบว่า RoomID ใหม่มีอยู่จริง
    var room entity.Rooms
    if err := db.First(&room, input.RoomID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบห้องที่ระบุ"})
        return
    }

    // ตรวจสอบว่า DriverID ใหม่มีอยู่จริง
    var driver entity.Driver
    if err := db.First(&driver, input.DriverID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ขับที่ระบุ"})
        return
    }

    // อัปเดตข้อมูลการจอง
    trainbook.Status = input.Status
    trainbook.RoomID = input.RoomID
    trainbook.DriverID = input.DriverID

    if err := db.Save(&trainbook).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถอัปเดตข้อมูลได้"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "อัปเดตการจองสำเร็จ", "trainbook": trainbook})
}

// ลบการจอง
func DeleteTrainBooking(c *gin.Context) {
	id := c.Param("id")
	var booking entity.TrainBook
	db := config.DB()
	if err := db.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลการจอง"})
		return
	}

	// ลดจำนวนการจองในห้อง
	var room entity.Rooms
	if err := db.First(&room, booking.RoomID).Error; err == nil {
		if room.CurrentBookings > 0 {
			room.CurrentBookings--
			db.Save(&room)
		}
	}

	// ลบข้อมูลการจอง
	if err := db.Delete(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลการจองได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ลบการจองสำเร็จ"})
}
